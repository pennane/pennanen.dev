import f from '../modules/frequencies.js'
import { squareWaveGenerator } from '../modules/waveform.js'

const song = {
    name: 'Take five',
    soundGenerator: squareWaveGenerator,
    frequencies: [f.AS2, f.DS3, f.DS4, [f.FS4, f.AS4, f.DS5], [f.AS3, f.F4, f.GS4, f.CS5], 0],
    step: 110,
    playingFunction: async (b, step, next) => {
        b[1].click()
        await next(2 * step)
        b[3].click()

        await next(1 * step)
        b[5].click()
        await next(2 * step)

        b[2].click()
        await next(1 * step)
        b[3].click()

        await next(1 * step)
        b[5].click()
        await next(2 * step)

        b[0].click()
        await next(3 * step)
        b[4].click()
        await next(3 * step)
    }
}

export default song
