import { createButton } from './main.js'
import { wait } from './util.js'

let song
let playing = false
let target
let finishedHook
const songs = [
    ['Pyramid Song', '../tracks/pyramid-song.js'],
    ['Everything in its right place', '../tracks/everything-in-its-right-place.js'],
    ['Take five', '../tracks/take-five.js']
]

async function next(time) {
    if (!playing) throw new Error()
    await wait(time)
    if (!playing) throw new Error()
}

async function unloadSong() {
    console.log('unloading')
    await new Promise((resolve, reject) => {
        if (playing) {
            finishedHook = resolve
            stopSong()
            playing = false
        } else {
            resolve()
        }
    })

    finishedHook = undefined
}

async function loadSong(_song, _target) {
    if (song) {
        await unloadSong()
    }

    const module = await import(_song[1])

    song = module.default
    target = _target

    while (target.firstChild) {
        target.removeChild(target.firstChild)
    }

    const container = document.createElement('div')

    song.frequencies.forEach((freq) => {
        container.appendChild(createButton(freq, song.soundGenerator))
    })
    target.appendChild(container)
}

async function loop() {
    const buttons = target.querySelectorAll('#target div button')
    await song.playingFunction(buttons, song.step, next)
    playing && (await loop())
}

async function playSong() {
    try {
        if (playing) return
        playing = true
        await loop()
    } catch {
        if (finishedHook) finishedHook()
    }
}

function stopSong() {
    playing = false
}

function loadedSongName() {
    return song ? song.name : null
}

export { loadSong, playSong, stopSong, songs, loadedSongName, unloadSong }
