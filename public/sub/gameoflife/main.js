let vh = window.innerHeight * 0.01
document.documentElement.style.setProperty('--vh', `${vh}px`)

const base = document.getElementById('game')
const container = document.querySelector('main')
const defaultCheckbox = document.getElementById('base-checkbox')
const startButton = document.getElementById('start-game')
const pauseButton = document.getElementById('pause-game')
const randomizeButton = document.getElementById('randomize-game')
const clearButton = document.getElementById('clear-game')

let { width, height } = getBoardSize()

let running = false
let board = []
let upcomingChanges = []

let targetFrameTime = 150

function getBoardSize() {
    return {
        width: Math.min(Math.floor(container.offsetWidth / (defaultCheckbox.offsetWidth + 1)) - 7, 64),
        height: Math.min(Math.floor(container.offsetHeight / (defaultCheckbox.offsetHeight + 1)) - 10, 64)
    }
}

function getRescaledBoard(b) {
    let board = [...b]
    const { width, height } = getBoardSize()

    if (board.length < width) {
        const initialLength = board.length
        board.length = width
        board.fill([], initialLength, width)
    } else {
        board = board.slice(0, width)
    }

    board.forEach((column) => {
        if (column.length < height) {
            const initialLength = column.length
            column.length = height
            column.fill(0, initialLength, height)
        } else {
            column = column.slice(0, height)
        }
    })

    return board
}

function getRandomCellValue() {
    return Math.random() > 0.5 ? 1 : 0
}

function getAliveNeighbours(board, x, y) {
    let amount = 0
    for (let cx = x - 1; cx <= x + 1; cx++) {
        if (cx < 0) continue
        if (cx > board.length - 1) continue

        for (let cy = y - 1; cy <= y + 1; cy++) {
            if (cy < 0) continue
            if (cy > board[0].length - 1) continue
            if (cy === y && x === cx) continue

            if (board[cx][cy] === 1) {
                amount++
            }
        }
    }
    return amount
}

function getNextCellState(board, x, y) {
    const cell = board[x][y]
    const alive = cell === 1
    const aliveNeighbours = getAliveNeighbours(board, x, y)

    if (aliveNeighbours === 3) {
        return 1
    }
    if (!alive) {
        return 0
    }

    if (aliveNeighbours === 2) {
        return 1
    }

    return 0
}

function getNextBoardState(board) {
    upcomingChanges.forEach(([x, y]) => {
        board[x][y] = board[x][y] === 1 ? 0 : 1
    })

    const nextBoard = []

    board.forEach((column, x) => {
        nextBoard[x] = []
        for (let y = 0; y < column.length; y++) {
            nextBoard[x][y] = getNextCellState(board, x, y)
        }
    })

    upcomingChanges = []

    return nextBoard
}

function createCheckbox(x, y) {
    const input = document.createElement('input')
    input.setAttribute('type', 'checkbox')
    input.setAttribute('data-x', x)
    input.setAttribute('data-y', y)

    return input
}

function createColumn() {
    const span = document.createElement('span')
    return span
}

function createBoard(defaultValue) {
    const board = []
    const { width, height } = getBoardSize()

    for (let i = 0; i < width; i++) {
        board[i] = []
        for (let j = 0; j < height; j++) {
            board[i][j] = defaultValue ?? getRandomCellValue()
        }
    }
    return board
}

function setButtonsToDOM(board) {
    while (base.childElementCount > board.length) {
        const last = base.lastChild
        base.removeChild(last)
    }
    while (base.childElementCount < board.length) {
        const button = createColumn()
        base.appendChild(button)
    }
    base.childNodes.forEach((c, x) => {
        while (c.childElementCount > board[0].length) {
            const last = c.lastChild
            c.removeChild(last)
        }
        if (c.childElementCount < board[0].length) {
            for (let y = c.childElementCount; y < board[0].length; y++) {
                const checkbox = createCheckbox(x, y)
                c.appendChild(checkbox)
            }
        }
    })
}

function rescale() {
    initialize(getRescaledBoard(board))
}

function displayBoard(board) {
    board.forEach((column, x) => {
        column.forEach((checked, y) => {
            const target = base.childNodes[x].childNodes[y]
            target.checked = checked === 1
        })
    })
}

let animationTimeout
function start() {
    clearTimeout(animationTimeout)
    running = true
    function animate() {
        board = getNextBoardState(board)
        displayBoard(board)

        if (running) {
            animationTimeout = setTimeout(() => requestAnimationFrame(() => animate()), targetFrameTime)
        }
    }
    requestAnimationFrame(() => animate())
}

function initialize(b, v) {
    running = false
    board = b || createBoard(v)
    setButtonsToDOM(board)
    displayBoard(board)
}

startButton.addEventListener('click', () => {
    start()
})

pauseButton.addEventListener('click', () => {
    running = false
})

clearButton.addEventListener('click', () => {
    initialize(undefined, 0)
})

randomizeButton.addEventListener('click', () => {
    initialize()
})

game.addEventListener('click', (e) => {
    if (e.target.nodeName === 'INPUT') {
        const x = Number(e.target.getAttribute('data-x'))
        const y = Number(e.target.getAttribute('data-y'))
        upcomingChanges.push([x, y])
    }
})

let timeout
window.addEventListener('resize', () => {
    clearTimeout(timeout)
    timeout = setTimeout(() => {
        let { width: updatedWidth, height: updatedHeight } = getBoardSize()
        if (updatedWidth - width !== 0 || updatedHeight - height !== 0) {
            width = updatedWidth
            height = updatedHeight
            rescale()
        }
    }, 250)
})

initialize()
