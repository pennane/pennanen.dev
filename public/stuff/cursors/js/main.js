const cursors = 
	  ["alias", "all-scroll", "auto", "cell", "context-menu", "col-resize", "copy", "crosshair", "default", "e-resize", "ew-resize", "grab", "grabbing", "help", "move", "n-resize", "ne-resize", "nesw-resize", "ns-resize", "nw-resize", "nwse-resize", "no-drop", "none", "not-allowed", "pointer", "progress", "row-resize", "s-resize", "se-resize", "sw-resize", "text", "url", "w-resize", "wait", "zoom-in", "zoom-out"]


let cursorsEl = document.createElement("div")
cursorsEl.id = "cursors"
cursors.forEach((cursor) => {
	let cursorEl = document.createElement("div")
	let cursorHead = document.createElement("h3")
	let cursorArea = document.createElement("div")
	
	cursorEl.className = "cursor parent"
	cursorHead.className = "cursor head"
	cursorArea.className = "cursor area " + cursor  
	cursorHead.textContent = cursor
	cursorEl.appendChild(cursorHead)
	cursorEl.appendChild(cursorArea)
	cursorsEl.appendChild(cursorEl)
})

document.querySelector("main").appendChild(cursorsEl)