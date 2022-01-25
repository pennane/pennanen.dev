import f from '../modules/frequencies.js'
import { squareWaveGenerator } from '../modules/waveform.js'

async function mainPart(b, step, next) {
    b[0].click()
    await next(step)
    b[1].click()
    await next(step)
    b[2].click()
    await next(step)
    b[3].click()
    await next(step)
    b[4].click()
    await next(3 * step)
    b[4].click()
    await next(2 * step)
    b[4].click()
    await next(2 * step)
    b[5].click()
    await next(2 * step)
    b[5].click()
    await next(2 * step)
    b[6].click()
    await next(2 * step)
    b[6].click()
    await next(3 * step)
}

async function otherPart(b, step, next) {
    b[7].click()
    await next(3 * step)
    b[8].click()
    await next(2 * step)
    b[8].click()
    await next(2 * step)
    b[5].click()
    await next(2 * step)
    b[5].click()
    await next(2 * step)
    b[6].click()
    await next(2 * step)
    b[6].click()
    await next(8 * step)
}

const song = {
    name: 'Everything in its right place',
    soundGenerator: squareWaveGenerator,
    frequencies: [
        f.C5,
        f.GS4,
        f.G4,
        f.C4,
        [f.C3, f.C4, f.E4, f.C5],
        [f.CS3, f.CS4, f.F4, f.C5],
        [f.DS3, f.DS4, f.G4, f.C5],
        [f.F3, f.C4, f.F4, f.A4, f.C5],
        [f.C3, f.E4, f.G4, f.C5]
    ],
    step: 240,
    playingFunction: async (b, step, next) => {
        await mainPart(b, step, next)
        await mainPart(b, step, next)
        await mainPart(b, step, next)
        await mainPart(b, step, next)
        await otherPart(b, step, next)
        await otherPart(b, step, next)
        await otherPart(b, step, next)
        await otherPart(b, step, next)
    }
}

export default song
