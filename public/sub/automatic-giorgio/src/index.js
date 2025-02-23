import { AudioContext } from 'standardized-audio-context'
import { default as musicLoader } from './musicloader.js'

const randomFromArray = arr => arr[Math.floor(Math.random() * arr.length)]
const randomFloatFromRange = (min, max) => Math.random() * (max - min + 1) + min
const randomIntegerFromRange = (min, max) =>
  Math.floor(randomFloatFromRange(min, max))

const audioCtx = new AudioContext()

let tracks = []
let loadedTracks = []

let tracknumber

const handleClick = async () => {
  try {
    tracks = await musicLoader()
  } catch (err) {
    console.log(err)
  }

  let source = audioCtx.createBufferSource()
  let currentTime = audioCtx.currentTime

  loadedTracks.slice(1).forEach(source => {
    source.stop()
  })

  loadedTracks = loadedTracks.slice(0, 1)

  let newTracknumber
  do {
    newTracknumber = randomIntegerFromRange(0, tracks.length - 1)
  } while (newTracknumber === tracknumber && tracks.length > 1)

  loadedTracks[1]

  tracknumber = newTracknumber
  source.buffer = tracks[tracknumber]
  source.loop = true
  source.connect(audioCtx.destination)
  source.start(0, currentTime)
  loadedTracks.unshift(source)
  newRandomAnimation()
}

class ResponsiveCanvas {
  constructor(
    target,
    options = { resize: true, fromOffset: true, translateToCenter: true }
  ) {
    this.canvas = target
    this.coordinates = [0, 0]
    this.ctx = target.getContext('2d')

    this.translateToCenter = options.translateToCenter

    if (options.fromOffset) {
      this.canvas.width = target.offsetWidth
      this.canvas.height = target.offsetHeight
    } else {
      this.canvas.width = options.width | target.width
      this.canvas.height = options.height | target.height
    }

    if (options.mousemove) {
      this.canvas.addEventListener('mousemove', e => {
        this.setCoordinates(e)
      })
    }
    if (options.resize) {
      this.scheduled = null
      window.addEventListener('resize', e => {
        if (!this.scheduled) {
          setTimeout(() => {
            this.resize()
            this.scheduled = null
          }, 500)
        }
        this.scheduled = true
      })
    }

    if (this.translateToCenter) {
      let transX = this.canvas.width * 0.5,
        transY = this.canvas.height * 0.5
      this.ctx.translate(transX, transY)
    }
  }
  getMousePos(event) {
    let rect = this.canvas.getBoundingClientRect()
    let scaleX = this.canvas.width / rect.width
    let scaleY = this.canvas.height / rect.height
    let coordinates = [
      (event.clientX - rect.left) * scaleX,
      (event.clientY - rect.top) * scaleY
    ]
    return coordinates
  }
  setCoordinates(event) {
    this.coordinates = this.getMousePos(event)
  }
  resize() {
    this.canvas.width = this.canvas.offsetWidth
    this.canvas.height = this.canvas.offsetHeight
    if (this.translateToCenter) {
      let transX = this.canvas.width * 0.5,
        transY = this.canvas.height * 0.5
      this.ctx.translate(transX, transY)
    }
  }
}

let target = document.getElementById('automatic-giorgio')
let responsivecanvas = new ResponsiveCanvas(target, {
  fromOffset: true,
  resize: true,
  translateToCenter: true
})

target.addEventListener('click', () => {
  handleClick()
})

let { canvas, ctx } = responsivecanvas

ctx.lineWidth = 2
window.points = 20
window.changeIntensity = 0.2
window.drawSize = 20
window.yFunction = 'sin'
window.xFunction = 'cos'
window.yMult = 1
window.xMult = 1

let animate

const generateXposition = val => {
  return Math[xFunction](val) * drawSize * xMult
}

const generateYposition = val => {
  return Math[yFunction](val) * drawSize * yMult
}

const getFPS = () => {
  return new Promise(resolve => {
    requestAnimationFrame(timeFrame1 => {
      requestAnimationFrame(timeFrame2 => {
        resolve(1000 / (timeFrame2 - timeFrame1))
      })
    })
  })
}

//Static variables, dont touch
window.fps = getFPS()
let rectSize = 4.5
let lineColor = 'rgb(0,0,0)'
let num = Math.floor(Math.random() * (360 - 0 + 1))
let dx, dy, odx, ody
let pointStore = []
//-------------------------------

const HslGen = () => {
  num += 0.1
  if (num > 360) {
    num = num - 360
  }
  return `hsl(${num}, 100%, 40%)`
}

const clearCanvas = () => {
  ctx.clearRect(
    -canvas.width / 2,
    -canvas.height / 2,
    canvas.width,
    canvas.height
  )
}

const drawdot = (gdx, gdy) => {
  clearCanvas()
  dx = gdx * rectSize
  dy = gdy * rectSize
  if (pointStore.length + 1 > points) {
    pointStore.shift()
    pointStore.push([dx, dy])
  } else {
    pointStore.push([dx, dy])
  }

  if (pointStore.length > points) {
    while (true) {
      pointStore.shift()
      if (!(pointStore.length > points)) {
        break
      }
    }
  }
  ctx.strokeStyle = lineColor
  ctx.moveTo(pointStore[0][0], pointStore[0][1])
  ctx.beginPath()
  pointStore.forEach((point, i) => {
    if (i === 0) return

    let tx = point[0] - pointStore[i - 1][0]
    let ty = point[1] - pointStore[i - 1][1]
    if (tx > 1000 || tx < -1000) {
      odx = dx
      return
    }
    if (ty > 1000 || ty < -1000) {
      ody = dy
      return
    }
    ctx.moveTo(pointStore[i - 1][0], pointStore[i - 1][1])
    ctx.lineTo(point[0], point[1])
  })
  ctx.stroke()
  ctx.closePath()
}
// --------------------

// Animator
let i = 0
animate = setInterval(
  (function x() {
    drawdot(
      generateXposition(i),
      generateYposition(i),
      canvas.width / 2,
      canvas.height / 2
    )
    i += changeIntensity
    lineColor = HslGen()
    return x
  })(),
  1000 / fps
)
// -----------------

let functionTypes = ['sin', 'cos', 'tan']

const newRandomAnimation = () => {
  xFunction = randomFromArray(functionTypes)
  do {
    yFunction = randomFromArray(functionTypes)
  } while (yFunction === xFunction)
  changeIntensity = randomFloatFromRange(0.2, 80)
  drawSize = randomIntegerFromRange(30, 100)
  points = randomIntegerFromRange(5, 42)
  //  point
  console.log(`current values:\n
    intensity: ${changeIntensity}\n
    drawSize: ${drawSize}\n
    points: ${points}\n
    mapped x function: ${xFunction}\n
    mapped y function: ${yFunction}\n
    `)
}
