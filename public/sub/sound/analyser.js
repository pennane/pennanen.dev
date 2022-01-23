import { mapBetween } from './util.js'
import { start, gap, amount } from './main.js'

let analyser, data
let animating = false

export function useAnalyzer(audioAnalyser) {
    data = new Uint8Array(audioAnalyser.frequencyBinCount)
    analyser = audioAnalyser
    console.log(analyser)
}

const canvas = document.getElementById('visualizer')
const context = canvas.getContext('2d')

canvas.width = canvas.offsetWidth
canvas.height = canvas.offsetHeight

new ResizeObserver((entries) => {
    const { width, height } = entries[0].contentRect
    canvas.width = width
    canvas.height = height
}).observe(canvas)

function draw(data) {
    context.clearRect(0, 0, canvas.width, canvas.height)
    const startValue = Math.floor(mapBetween(Math.max(start - 2 * gap, 0), 0, 22000, 0, data.length - 1))
    const endValue = Math.floor(
        mapBetween(Math.min(start + amount * gap + 2 * gap, 22000), 0, 22000, 0, data.length - 1)
    )

    const space = canvas.width / (endValue - startValue)

    context.beginPath()
    context.moveTo(0, canvas.height)

    let line = 1
    for (let i = startValue; i < endValue; i++) {
        const value = data[i]
        context.lineTo(space * line, canvas.height - value)
        line++
    }
    context.lineTo(canvas.width, canvas.height)
    context.stroke()
}

function animate() {
    if (!analyser) return
    analyser.getByteFrequencyData(data)
    draw(data)
    if (animating) requestAnimationFrame(animate)
}

export function startAnalysing() {
    animating = true
    requestAnimationFrame(animate)
}
export function stopAnalysing() {
    animating = false
}
