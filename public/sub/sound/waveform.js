export function* sineWave(audioContext, frequency) {
    const sampleFrequency = audioContext.sampleRate / frequency
    let i = 0
    while (true) {
        i++
        yield Math.sin(i / (sampleFrequency / (Math.PI * 2)))
    }
}

export function* squareWave(audioContext, frequency) {
    const sampleFrequency = audioContext.sampleRate / frequency
    let i = 0
    while (true) {
        i++
        yield Math.round(Math.sin(i / (sampleFrequency / (Math.PI * 2))))
    }
}

export function whiteNoise(audioContext) {
    return new Float32Array(audioContext.sampleRate).map(() => Math.random())
}
