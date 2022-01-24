import { mapBetween } from './util.js'
import { start, gap, amount, createHoverSource } from './main.js'

let analyser, data
let animating = false
let oldBinCount, oldStart, oldAmount, oldGap, startHz, endHz

let startValue
let endValue

function updateDataArray() {
    data = new Uint8Array(analyser.frequencyBinCount)
}

export function useAnalyzer(audioAnalyser) {
    analyser = audioAnalyser
    updateDataArray()
}

const canvas = document.getElementById('visualiser')
const context = canvas.getContext('2d')

canvas.width = canvas.offsetWidth
canvas.height = canvas.offsetHeight

context.font = '20px Arial'

new ResizeObserver((entries) => {
    const { width, height } = entries[0].contentRect
    canvas.width = width
    canvas.height = height
}).observe(canvas)

function clearCanvas() {
    context.clearRect(0, 0, canvas.width, canvas.height)
}

function draw(data, startValue, endValue) {
    const space = canvas.width / (endValue - startValue)

    clearCanvas()

    for (let i = 0; i < 10; i++) {
        let x = (canvas.width / 10) * i
        context.fillStyle = '#FF0000'
        context.fillRect(x - 1, 0, 1, canvas.height)
        context.fillStyle = '#000'
        context.fillText(`${frequencyAtX(x)} hz`, x + 5, canvas.height - 8)
        frequencyAtX
    }

    context.fillStyle = '#000000'

    context.beginPath()
    context.moveTo(0, canvas.height)

    let line = 1

    if (canvas.height < 300) {
        for (let i = startValue; i < endValue; i++) {
            const value = mapBetween(data[i], 0, 255, 0, canvas.height - 10)
            context.lineTo(space * line, canvas.height - value)
            line++
        }
    } else {
        for (let i = startValue; i < endValue; i++) {
            const value = data[i]
            context.lineTo(space * line, canvas.height - value)
            line++
        }
    }

    context.lineTo(canvas.width, canvas.height)
    context.stroke()
}

function animate() {
    if (!analyser) return
    let bincountChanged = false
    if (!oldBinCount || oldBinCount !== analyser.frequencyBinCount) {
        bincountChanged = true
        updateDataArray()
        oldBinCount = analyser.frequencyBinCount
    }

    if (oldStart !== start || oldGap !== gap) {
        startHz = start
        endHz = start + amount * gap
    }

    if (bincountChanged || !startValue || !endValue || start !== oldStart || gap !== oldGap || amount !== oldAmount) {
        startValue = Math.floor(mapBetween(startHz, 0, 22000, 0, data.length - 1))
        endValue = Math.floor(mapBetween(endHz, 0, 22000, 0, data.length - 1))

        oldStart = start
        oldGap = gap
        oldAmount = amount
    }

    analyser.getByteFrequencyData(data)
    draw(data, startValue, endValue)
    if (animating) requestAnimationFrame(animate)
    else clearCanvas()
}

function frequencyAtX(x) {
    return Math.round(mapBetween(x, 0, canvas.offsetWidth, startHz, endHz))
}

canvas.addEventListener('click', (e) => {
    e.preventDefault()
    if (!isAnalysing()) return
    const x = e.clientX - canvas.offsetLeft
    createHoverSource(frequencyAtX(x))
})

function startAnalysing() {
    animating = true
    requestAnimationFrame(animate)
}
function stopAnalysing() {
    animating = false
    clearCanvas()
}
function isAnalysing() {
    return !!animating
}

export { startAnalysing, stopAnalysing, isAnalysing }
