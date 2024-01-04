const tag = "hlgl";

document.addEventListener("DOMContentLoaded", function () {
    function generateRandomString(length) {
        const characters =
            "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789";
        let result = "";
        for (let i = 0; i < length; i++) {
            result += characters.charAt(
                Math.floor(Math.random() * characters.length)
            );
        }
        return result;
    }

    let elementsWithStyle = document.querySelectorAll("[style]"),
        dynamicStylesheet = document.createElement("style");

    dynamicStylesheet.setAttribute(tag, "sheet");
    document.head.appendChild(dynamicStylesheet);

    elementsWithStyle.forEach((element) => {
        let randomString = generateRandomString(3);

        while (document.querySelector(`[${tag}="${randomString}"]`)) {
            randomString = generateRandomString(3);
        }

        element.setAttribute(tag, randomString);

        let currentStyles = dynamicStylesheet.innerText;
        dynamicStylesheet.textContent =
            currentStyles +
            `[${tag}="${randomString}"]{${element.getAttribute("style")}}`;

        element.removeAttribute("style");
    });

    let stylesheetContent = dynamicStylesheet.textContent;

    const replacements = [
        {
            pattern: /[\n\t ]{4,}/g,
            replacement: "",
        },
        {
            pattern: /@dark/g,
            replacement: "@media (prefers-color-scheme: dark)",
        },
        {
            pattern: /@(\d+)px\-/g,
            replacement: "@media (max-width: $1px)",
        },
        {
            pattern: /@(\d+)px\+/g,
            replacement: "@media (min-width: $1px)",
        },
    ];

    replacements.forEach(({ pattern, replacement }) => {
        stylesheetContent = stylesheetContent.replace(pattern, replacement);
    });

    dynamicStylesheet.textContent = stylesheetContent;
});
