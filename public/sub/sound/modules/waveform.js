export function* sineWaveGenerator(audioContext, frequency) {
    const sampleFrequency = audioContext.sampleRate / frequency
    let i = 0
    while (true) {
        i++
        yield Math.sin(i / (sampleFrequency / (Math.PI * 2)))
    }
}

export function* squareWaveGenerator(audioContext, frequency) {
    const sampleFrequency = audioContext.sampleRate / frequency
    let i = 0
    while (true) {
        i++
        yield Math.floor(Math.sin(i / (sampleFrequency / (Math.PI * 2))))
    }
}

export function createWhiteNoise(audioContext) {
    return new Float32Array(audioContext.sampleRate).map(() => Math.random())
}
