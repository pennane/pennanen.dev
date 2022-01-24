import f from '../modules/frequencies.js'
import { squareWaveGenerator } from '../modules/waveform.js'

const song = {
    name: 'Pyramid song',
    soundGenerator: squareWaveGenerator,
    frequencies: [
        [f.FS2, f.CS3, f.FS3, f.AS3, f.CS4, f.FS4],
        [f.G2, f.B2, f.D3, f.G3, f.B3, f.FS4],
        [f.A2, f.E3, f.G3, f.CS4, f.FS4],
        [f.G2, f.B2, f.D3, f.G3, f.B3, f.G4]
    ],
    step: 100,
    playingFunction: async (b, step, next) => {
        b[0].click()
        await next(8 * step)
        b[0].click()
        await next(8 * step)
        b[1].click()
        await next(12 * step)
        b[2].click()
        await next(8 * step)
        b[2].click()
        await next(8 * step)
        b[2].click()
        await next(8 * step)
        b[1].click()
        await next(8 * step)
        b[1].click()
        await next(12 * step)
        b[3].click()
        await next(8 * step)
        b[3].click()
        await next(8 * step)
    }
}

export default song
