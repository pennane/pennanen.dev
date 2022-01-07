let head = document.getElementsByTagName("head")[0],
	link = document.createElement("link");
link.rel = "stylesheet",
	link.type = "text/css",
	link.href = "style/night.css",
	link.id = "darktheme",
	link.disabled = !0,
	head.appendChild(link);
let darkThemeElement = document.getElementById(link.id);

(function () {
	if (window.location.hash.toLowerCase() === "#dark") loadDarkTheme()
})()

function loadDarkTheme() {
	darkThemeElement.disabled = !1
}