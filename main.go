package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	jsoniter "github.com/json-iterator/go"
	"google.golang.org/api/option"

	"github.com/gofiber/fiber/v2"

	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/helmet"

	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/earlydata"
)

var json = jsoniter.ConfigFastest

var htmlre *regexp.Regexp
var cssre *regexp.Regexp
var jsre *regexp.Regexp

func makeCodeblockRe(lang string) *regexp.Regexp {
	return regexp.MustCompile(fmt.Sprintf(`(?m)^\x60{3}%v\n([\s\S]*?)\n\x60{3}$`, lang))
}

func compileCodeblocks(raw string) string {
	cde := htmlre.FindStringSubmatch(raw)
	if len(cde) == 0 {
		return raw
	}
	code := cde[1]
	cde = strings.Split(code, "</body>")
	if len(cde) == 0 {
		return raw
	}
	for _, scripts := range jsre.FindAllStringSubmatch(raw, -1) {
		cde[len(cde)-2] += "<script>" + scripts[1] + "</script>"
	}

	for _, css := range cssre.FindAllStringSubmatch(raw, -1) {
		cde[len(cde)-2] += "<style>" + css[1] + "</style>"
	}
	return strings.Join(cde, "</body>")
}

// keep order
func RemoveO[T any](s []T, i int) []T {
	return append(s[:i], s[i+1:]...)
}

type Search struct {
	Name        string
	Description string
}

var model *genai.GenerativeModel
var ctx context.Context
var searchParser *regexp.Regexp

func genSearchResults(prompt string) (res []Search) {
	resp, err := model.GenerateContent(ctx, genai.Text(fmt.Sprintf(`You must generate 10 made up search results with names and descriptions of possible webpages about the topic encased in ", each search result must be in this format: 

&&&& Name: (a name goes here, it must be less than 32 characters)
Description: (description goes here, it must be less than 128 characters)

Here is the prompt: "%v". Do not present the results as an ordered or unordered list. Only reply with text, no images or formatting to text.`, prompt)))

	if err != nil {
		log.Printf("error in genSearchResults(%v): %v", prompt, err.Error())
		return []Search{{Name: "An error!", Description: "We are sorry! Please try again"}}
	}

	for _, v := range searchParser.FindAllStringSubmatch(string(resp.Candidates[0].Content.Parts[0].(genai.Text)), -1) {
		res = append(res, Search{Name: v[1], Description: v[2]})
	}
	return
}

func genWebPage(search Search) string {
	resp, err := model.GenerateContent(ctx, genai.Text(fmt.Sprintf(`You must only generate an HTML, Javascript and CSS webpage code with the following description and name that is encased in ", also make sure that the CSS theme matches the subject of the site, for example: use darker colors for darker subjects, truly go wild with the styling and colors, and make the css feel like it is matching the subject, when making <img>s, set the source to https://source.unsplash.com/featured/?[replace this with what you want the image to display in one word], you must make the website look more modern and include all sorts of interesting scripts, and styles related to the subject. You must not use the backtick character in the response. Only respond with the code. 
  
It must be a single HTML file, javascript and css must be fully included in the HTML code, not otherwise. What you send must be the raw code and will be put into the index.html file, creating other files is not allowed. Splitting the code to multiple files is not allowed.
Name: "%v"
Description: "%v"`, search.Name, search.Description)))

	if err != nil {
		log.Printf("error in genWebPage(%v): %v", search, err.Error())
		return "An error has occured during webpage generation :("
	}

	return string(resp.Candidates[0].Content.Parts[0].(genai.Text))
}

func main() {
	searchParser = regexp.MustCompile(`(?m)&&&& Name: (.+)
Description: (.+)`)
	htmlre = makeCodeblockRe("html5?")
	cssre = makeCodeblockRe("css3?")
	jsre = makeCodeblockRe(`(?:javascript|js)`)
	ctx = context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("api")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	model = client.GenerativeModel("gemini-pro")

	app := fiber.New(fiber.Config{
		Prefork:     true,
		JSONEncoder: json.Marshal,
		JSONDecoder: json.Unmarshal,
	})

	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))
	app.Use(earlydata.New())
	app.Use(recover.New())
	app.Use(helmet.New(helmet.Config{CrossOriginEmbedderPolicy: "credentialless", CrossOriginResourcePolicy: "cross-origin"}))
	app.Use(func(c *fiber.Ctx) error {
		c.Response().Header.Del("X-Frame-Options")
		return c.Next()
	})
	app.Use(cors.New())
	//app.Use(limiter.New())

	app.Get("/api/searchResults", func(c *fiber.Ctx) error {
		prompt := c.Query("prompt")
		if prompt == "" {
			return c.SendStatus(fiber.StatusNoContent)
		}

		res := genSearchResults(prompt)
		return c.JSON(res)
	})

	app.Get("/api/webPage", func(c *fiber.Ctx) error {
		name := c.Query("name")
		description := c.Query("description")
		if name == "" || description == "" {
			return c.SendStatus(fiber.StatusBadRequest)
		}

		// split := strings.Split(res, "\n")
		// if split[0] != "<!DOCTYPE html>" { // combat the fucking ai sending codeblocks (fuck you)
		//   split = RemoveO(split, 0)
		//   split = RemoveO(split, len(split)-1)
		//   res = strings.Join(split, "\n")
		// }

		c.Set("Content-Type", "text/html")
		return c.SendString(compileCodeblocks(genWebPage(Search{Name: name, Description: description})))
	})

	app.Static("/", "/data/served/", fiber.Static{
		Compress:      true,
		CacheDuration: 4 * time.Hour,
		MaxAge:        4 * 60 * 60, // 4hrs
	})

	log.Fatal(app.Listen(":4664"))
}
