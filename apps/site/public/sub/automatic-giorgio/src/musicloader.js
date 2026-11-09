import { AudioContext } from 'standardized-audio-context'

const testAudio = document.createElement('audio')
const canPlayOgg =
  typeof testAudio.canPlayType === 'function' &&
  testAudio.canPlayType('audio/ogg') !== ''

const fileType = canPlayOgg ? '.ogg' : '.mp3'

const randomFloatFromRange = (min, max) => Math.random() * (max - min + 1) + min
const randomIntegerFromRange = (min, max) =>
  Math.floor(randomFloatFromRange(min, max))

const loadAudioBuffer = (url, context) => {
  return new Promise((resolve, reject) => {
    fetch(url)
      .then(response => response.arrayBuffer())
      .then(arrayBuffer => context.decodeAudioData(arrayBuffer))
      .then(audioBuffer => {
        return resolve(audioBuffer)
      })
      .catch(err => reject(err))
  })
}

const loadTrack = id => {
  return new Promise(async (resolve, reject) => {
    const url = './assets/music/giorgio/g' + id + fileType
    const context = new AudioContext()
    let audio = await loadAudioBuffer(url, context)
    return resolve(audio)
  })
}

const amountOfTracks = 11

let trackIds = []
let tracks = []
let trackPromises = []

for (let i = 0; i < amountOfTracks; i++) {
  let id
  do {
    id = randomIntegerFromRange(0, 106)
  } while (trackIds.indexOf(id) > -1)
  trackIds.push(id)
}

async function loadToMemory() {
  for (const trackid of trackIds) {
    let track = loadTrack(trackid)
    trackPromises.push(track)
    let resolvedTrack = await track
    tracks.push(resolvedTrack)
  }
}

loadToMemory()

const getTracks = () => {
  return new Promise(async (resolve, reject) => {
    if (tracks.length === 0) {
      Promise.any(trackPromises)
        .then(() => {
          return resolve(tracks)
        })
        .catch(err => reject(err))
    } else {
      return resolve(tracks)
    }
  })
}

export default getTracks
