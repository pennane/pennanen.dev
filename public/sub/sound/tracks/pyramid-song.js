import f from '../modules/frequencies.js'
import { sineWaveGenerator, squareWaveGenerator } from '../modules/waveform.js'

const fsGmajA6Gmaj = async (b, step, next) => {
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
}

const fsmEaddGmaj = async (b, step, next) => {
    b[4].click()
    await next(14 * step)
    b[5].click()
    await next(20 * step)
    b[1].click()
    await next(16 * step)
}

const fsGmajA6GmajG = async (b, step, next) => {
    await fsGmajA6Gmaj(b, step, next)
    b[3].click()
    await next(8 * step)
    b[3].click()
    await next(8 * step)
}

const fsGmajA6GmajFs = async (b, step, next) => {
    await fsGmajA6Gmaj(b, step, next)
    b[0].click()
    await next(8 * step)
    b[0].click()
    await next(8 * step)
}

const intro = async (b, step, next) => {
    await fsGmajA6GmajG(b, step, next)
    await fsGmajA6Gmaj(b, step, next)
    await fsGmajA6Gmaj(b, step, next)
    await fsGmajA6Gmaj(b, step, next)
}

const verse = async (b, step, next) => {
    await fsGmajA6GmajFs(b, step, next)
    await fsmEaddGmaj(b, step, next)
    await fsGmajA6GmajFs(b, step, next)
    await fsmEaddGmaj(b, step, next)
    await fsGmajA6GmajFs(b, step, next)
    await fsmEaddGmaj(b, step, next)
}

const song = {
    name: 'Pyramid song',
    soundGenerator: squareWaveGenerator,
    frequencies: [
        [f.FS2, f.CS3, f.FS3, f.AS3, f.CS4, f.FS4],
        [f.G2, f.B2, f.D3, f.G3, f.B3, f.FS4],
        [f.A2, f.E3, f.G3, f.CS4, f.FS4],
        [f.G2, f.B2, f.D3, f.G3, f.B3, f.G4],
        [f.FS2, f.CS3, f.FS3, f.A3, f.CS4, f.FS4],
        [f.E2, f.GS2, f.B2, f.FS3],
        [f.FS2, f.CS3, f.FS3, f.AS3, f.E4, f.G4]
    ],
    step: 110,
    playingFunction: async (b, step, next) => {
        await intro(b, step, next)
        await verse(b, step, next)
    }
}

export default song
