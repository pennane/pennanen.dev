let base = document.getElementById('game-of-life')

let size = {
    x: 0,
    y: 0
}

let updateInterval = 200;

function createBlock(x, y) {
    let block = document.createElement('input')
    block.type = "checkbox"
    block.name = 'block' + x + '-' + y
    block.id = 'block' + x + '-' + y
    return block
}

let baseCheckbox = document.getElementById('base-checkbox')
let boxWidth = parseInt(getComputedStyle(baseCheckbox).width) + 3
let boxHeight = parseInt(getComputedStyle(baseCheckbox).height) + 8

baseCheckbox.style.display = 'none'

function scaleGameBoard() {
    let gameDiv = document.getElementById('game')
    let gameWidth = parseInt(getComputedStyle(gameDiv).width)
    let gameHeight = parseInt(getComputedStyle(gameDiv).height)

    let buttons = document.getElementById('buttons')
    let buttonsHeight = parseInt(getComputedStyle(buttons).height)
    size = {
        x: Math.floor(gameWidth / boxWidth),
        y: Math.floor((gameHeight - buttonsHeight) / boxHeight)
    };

    [...document.querySelectorAll('#gameboard')].forEach(div => {
        div.parentNode.removeChild(div);
    })

    let gameBoard = document.createElement('div')
    gameBoard.id = "gameboard"

    for (let y = 0; y <= size.y; y++) {
        let row = document.createElement('div')
        row.id = "row-" + y
        for (let x = 0; x <= size.x; x++) {
            let block = createBlock(x, y)
            row.appendChild(block)
        }
        gameBoard.appendChild(row)
    }

    base.appendChild(gameBoard)
}

scaleGameBoard()

let running = false;

function checkNeigbhours(x, y) {
    let alive = 0;
    for (let nx = x - 1; nx <= x + 1; nx++) {
        for (let ny = y - 1; ny <= y + 1; ny++) {
            if (ny < 0 || nx < 0) {
                continue
            }
            if (nx > size.x || ny > size.y) {
                continue
            }
            if (nx === x && ny === y) {
                continue
            }

            let name = 'block' + nx + '-' + ny
            let state = document.getElementById(name).checked
            if (state) alive++;
        }
    }
    return alive
}

function randomizeGameBoard() {
    for (let x = 0; x <= size.x; x++) {
        for (let y = 0; y <= size.y; y++) {
            let block = document.getElementById('block' + x + '-' + y)
            block.checked = Math.random() >= 0.7
        }
    }
}

function clearGameBoard() {
    for (let x = 0; x <= size.x; x++) {
        for (let y = 0; y <= size.y; y++) {
            let block = document.getElementById('block' + x + '-' + y)
            block.checked = false
        }
    }
}

function calculateGameState() {
    let state = []
    for (let x = 0; x <= size.x; x++) {
        state[x] = []
        for (let y = 0; y <= size.y; y++) {
            let block = document.getElementById('block' + x + '-' + y)

            let alive = block.checked
            let neighbours = checkNeigbhours(x, y)

            state[x][y] = alive;

            if (alive && neighbours < 2) {
                state[x][y] = false
            } else if (alive && (neighbours === 2 || neighbours === 3)) {
                state[x][y] = true
            } else if (alive && neighbours > 3) {
                state[x][y] = false
            } else if (!alive && neighbours === 3) {
                state[x][y] = true
            }
        }
    }
    for (let x = 0; x <= size.x; x++) {
        for (let y = 0; y <= size.y; y++) {
            let block = document.getElementById('block' + x + '-' + y)
            block.checked = state[x][y]
        }
    }
    if (running) {
        setTimeout(() => { calculateGameState() }, updateInterval)
    }

}

document.getElementById('start-game').addEventListener('click', () => {
    running = true;
    calculateGameState()
})

document.getElementById('pause-game').addEventListener('click', () => {
    running = false;
})

document.getElementById('randomize-game').addEventListener('click', () => {
    randomizeGameBoard()
})

document.getElementById('clear-game').addEventListener('click', () => {
    running = false
    clearGameBoard()
})

document.getElementById('rescale-game').addEventListener('click', () => {
    running = false
    clearGameBoard()
    scaleGameBoard()
})