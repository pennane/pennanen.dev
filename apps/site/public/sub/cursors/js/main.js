const cursors = [
    'alias',
    'all-scroll',
    'auto',
    'cell',
    'context-menu',
    'col-resize',
    'copy',
    'crosshair',
    'default',
    'e-resize',
    'ew-resize',
    'grab',
    'grabbing',
    'help',
    'move',
    'n-resize',
    'ne-resize',
    'nesw-resize',
    'ns-resize',
    'nw-resize',
    'nwse-resize',
    'no-drop',
    'none',
    'not-allowed',
    'pointer',
    'progress',
    'row-resize',
    's-resize',
    'se-resize',
    'sw-resize',
    'text',
    'url',
    'w-resize',
    'wait',
    'zoom-in',
    'zoom-out'
]

const container = document.createElement('div')
container.id = 'cursors'

for (const cursor of cursors) {
    const cursorElement = document.createElement('div')
    cursorElement.className = 'cursor parent'

    const cursorHeader = document.createElement('h3')
    cursorHeader.className = 'cursor head'
    cursorHeader.textContent = cursor

    const cursorBody = document.createElement('div')
    cursorBody.className = 'cursor area ' + cursor

    cursorElement.appendChild(cursorHeader)
    cursorElement.appendChild(cursorBody)

    container.appendChild(cursorElement)
}

document.querySelector('main').appendChild(container)
