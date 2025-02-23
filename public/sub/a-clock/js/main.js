const size = 800
const canvas = document.createElement('canvas')
canvas.height = size
canvas.width = size

const ctx = canvas.getContext('2d')
ctx.translate(canvas.width / 2, canvas.height / 2)

document.getElementById('canvas-parent').appendChild(canvas)

const getSummedTime = () => {
  let now = new Date()
  let time = {
    hours: now.getHours() % 12,
    minutes: now.getMinutes(),
    seconds: now.getSeconds(),
    mseconds: now.getMilliseconds(),
  }
  return {
    hours:
      (time.hours / 12) * 360 +
      (time.minutes / 60 / 12) * 360 +
      (time.seconds / 60 / 60 / 12) * 360 +
      (time.mseconds / 1000 / 60 / 60 / 12) * 360,
    minutes:
      (time.minutes / 60) * 360 +
      (time.seconds / 60 / 60) * 360 +
      (time.mseconds / 1000 / 60 / 60) * 360,
    seconds: (time.seconds / 60) * 360 + (time.mseconds / 1000 / 60) * 360,
  }
}

const rotatePoint = ([x, y], deg) => {
  const radians = (deg * Math.PI) / 180

  return [
    x * Math.cos(radians) - y * Math.sin(radians),
    x * Math.sin(radians) + y * Math.cos(radians),
  ]
}

const setup = () => {
  ctx.fillStyle = 'white'
  ctx.strokeStyle = 'white'
  ctx.lineWidth = 10
  ctx.textAlign = 'center'
  ctx.textBaseline = 'middle'

  requestAnimationFrame(draw)
}

const drawLine = (width, start, length, angle) => {
  ctx.beginPath()
  ctx.moveTo(...rotatePoint([0, start], angle))
  ctx.lineWidth = width
  ctx.lineTo(...rotatePoint([0, start + length], angle))
  ctx.stroke()
}

const drawCircle = (size, dist, angle) => {
  ctx.beginPath()
  ctx.arc(...rotatePoint([0, dist], angle), size, 0, 2 * Math.PI)
  ctx.fill()
}

const drawRotatedText = (text, size, dist, angle, invert = false) => {
  ctx.save()
  ctx.translate(...rotatePoint([0, dist], angle + 180))

  // Apply additional 180 degrees rotation if invert is true
  const textRotationAngle = invert ? angle + 180 : angle
  ctx.rotate((textRotationAngle * Math.PI) / 180)

  ctx.font = `${size}px 'Inter', serif`
  ctx.fillText(text, 0, 0)
  ctx.restore()
}

const drawText = (text, size, dist, angle) => {
  ctx.font = `${size}px 'Inter', serif`

  ctx.fillText(text, ...rotatePoint([0, dist], angle + 180))
  ctx.restore()
}

const drawFace = () => {
  for (let i = 0; i < 240; i++) {
    if (i % 20 == 0) {
      drawCircle(5.5, size / 2 - 22, (360 / 240) * i)
    } else if (i % 4 == 0) {
      drawLine(3, size / 2 - 50, 35, (360 / 240) * i)
    } else {
      drawLine(2, size / 2 - 32, 17, (360 / 240) * i)
    }
  }

  for (let i = 1; i <= 12; i++) {
    drawRotatedText(i * 5, 30, size / 2 - 49, (360 / 12) * i, i > 4 && i < 9)

    if (i === 3) {
      ctx.fillStyle = 'white'
      ctx.rect(size / 3 - 1, -27, 57, 46)
      ctx.fill()

      ctx.fillStyle = 'black'
      ctx.font = `45px lighter 'Inter', serif`
      ctx.fillText(new Date().getDate(), size / 3 + 25, 0)
      ctx.fillStyle = 'white'
      ctx.rect(size / 3 - 8, -35, 70, 61)
      ctx.stroke()
    } else {
      drawText(i, 90, size / 2 - 112, (360 / 12) * i)
    }

    drawText(i + 12, 40, size / 3 - 59, (360 / 12) * i)
  }

  ctx.letterSpacing = '4px'
  ctx.font = `35px 'Inter', serif`

  ctx.fillText('HAMILTON', 0, -size / 7)

  ctx.font = `26px 'Inter', serif`
  ctx.fillText('KHAKI', 0, size / 8.5)

  ctx.font = `26px 'Inter', serif`
  ctx.fillText('AUTOMATIC', 0, size / 6.4)
  ctx.letterSpacing = '-2px'

  drawCircle(10, 0, 0)
}

const hoursHand = document.getElementById('hours')
const minutesHand = document.getElementById('minutes')
const secondsHand = document.getElementById('seconds')

const rotateHands = () => {
  const { hours, minutes, seconds } = getSummedTime()

  hoursHand.style.transform = `rotateZ(${hours}deg)`
  minutesHand.style.transform = `rotateZ(${minutes}deg)`
  secondsHand.style.transform = `rotateZ(${seconds}deg)`
}

const clear = ctx.clearRect.bind(ctx, -size / 2, -size / 2, size, size)

const draw = () => {
  clear()
  drawFace()
  rotateHands()
  requestAnimationFrame(draw)
}

setup()
