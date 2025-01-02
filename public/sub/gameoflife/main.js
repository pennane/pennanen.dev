// @ts-check
let vh = window.innerHeight * 0.01
document.documentElement.style.setProperty('--vh', `${vh}px`)

const toggle_game_state_element = document.getElementById('start-game')
const game_view_element = document.getElementById('game')
const container_element = document.querySelector('main')
const checkbox_prototype_element = document.getElementById('base-checkbox')

const MAX_SIZE = 64, TARGET_FRAME_TIME = 150;

let m_running = false
let m_board = []
let m_queue = []

function calculate_board_size() {
    return {
        width: Math.min(Math.floor(container_element.offsetWidth / (checkbox_prototype_element.offsetWidth + 1)) - 7, MAX_SIZE),
        height: Math.min(Math.floor(container_element.offsetHeight / (checkbox_prototype_element.offsetHeight + 1)) - 10, MAX_SIZE)
    }
}

function rescale_board(board, width, height) {
    return Array.from({ length: width }, (_, x) => Array.from({ length: height }, (_, y) => board[x]?.[y] ?? 0))
}

function random_cell_value() {
    return Math.round(Math.random())
}


const NEIGHBOR_DELTAS = [[0, 1], [1, 1], [1, 0], [1, -1], [0, -1], [-1, -1], [-1, 0], [-1, 1]]

function alive_neighbor_count(board, x, y) {
    return NEIGHBOR_DELTAS.reduce((count, [dx, dy]) => board[x + dx]?.[y + dy] ? count + 1 : count, 0)
}

function next_cell_state(board, x, y) {
    const cell = board[x][y]
    const alive = cell === 1
    const neighbors = alive_neighbor_count(board, x, y)

    if (neighbors === 3) {
        return 1
    }
    if (!alive) {
        return 0
    }

    if (neighbors === 2) {
        return 1
    }

    return 0
}

function inject_queue(board) {
    const newBoard = board.map(row => row.slice())
    m_queue.forEach(([x, y]) => {
        newBoard[x][y] = newBoard[x][y] === 1 ? 0 : 1
    })
    m_queue = []
    return newBoard
}

function next_board_state(board) {
    const nextBoard = []
    board.forEach((column, x) => {
        nextBoard[x] = []
        for (let y = 0; y < column.length; y++) {
            nextBoard[x][y] = next_cell_state(board, x, y)
        }
    })

    return nextBoard
}

function create_checkbox(x, y) {
    const input = document.createElement('input')
    input.setAttribute('type', 'checkbox')
    input.setAttribute('data-x', x)
    input.setAttribute('data-y', y)

    return input
}

function create_column() {
    const span = document.createElement('span')
    return span
}

function create_board(default_cell_value, width, height) {
    const board = []

    for (let i = 0; i < width; i++) {
        board[i] = []
        for (let j = 0; j < height; j++) {
            board[i][j] = default_cell_value ?? random_cell_value()
        }
    }
    return board
}

function insert_buttons(board, base) {
    while (base.childElementCount > board.length) {
        const last = base.lastChild
        base.removeChild(last)
    }
    while (base.childElementCount < board.length) {
        const button = create_column()
        base.appendChild(button)
    }
    base.childNodes.forEach((c, x) => {
        while (c.childElementCount > board[0].length) {
            const last = c.lastChild
            c.removeChild(last)
        }
        if (c.childElementCount < board[0].length) {
            for (let y = c.childElementCount; y < board[0].length; y++) {
                const checkbox = create_checkbox(x, y)
                c.appendChild(checkbox)
            }
        }
    })
}


function display_board(board, grid) {
    board.forEach((column, x) => {
        column.forEach((checked, y) => {
            const target = grid.childNodes[x].childNodes[y]
            target.checked = checked === 1
        })
    })
}

let animation_timeout
function start() {
    clearTimeout(animation_timeout)
    m_running = true
    function animate() {
        if (!m_running) return
        m_board = next_board_state(inject_queue(m_board))
        display_board(m_board, game_view_element)
        animation_timeout = setTimeout(animate, TARGET_FRAME_TIME)
    }
    animate()
}

function initialize(oldBoard, cellValue) {
    let board
    if (oldBoard) {
        board = oldBoard
    } else {
        const { width, height } = calculate_board_size()
        board = create_board(cellValue, width, height)
    }

    insert_buttons(board, game_view_element)
    display_board(board, game_view_element)

    m_board = board
}

function set_running(state) {
    m_running = state
    if (state) {
        start()
        toggle_game_state_element.textContent = "pause"
    } else {
        toggle_game_state_element.textContent = "start"
    }

}


toggle_game_state_element?.addEventListener('click', () => {
    set_running(!m_running)
})

document.getElementById('clear-game').addEventListener('click', () => {
    initialize(undefined, 0)
    set_running(false)
})

document.getElementById('randomize-game').addEventListener('click', () => {
    initialize()
})

game_view_element.addEventListener('click', (e) => {
    if (e.target.nodeName === 'INPUT') {
        const x = Number(e.target.getAttribute('data-x'))
        const y = Number(e.target.getAttribute('data-y'))
        m_queue.push([x, y])
    }
})

let resizing_debounce_timeout
window.addEventListener('resize', () => {
    clearTimeout(resizing_debounce_timeout)
    resizing_debounce_timeout = setTimeout(() => {
        const { width, height } = calculate_board_size()
        initialize(rescale_board(m_board, width, height))
    }, 20)
})

initialize()
