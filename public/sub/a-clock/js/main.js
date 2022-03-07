const hoursHand = document.getElementById('hours')
const minutesHand = document.getElementById('minutes')
const secondsHand = document.getElementById('seconds')

const getSummedTime = () => {
    let now = new Date()
    let time = {
        hours: now.getHours() % 12,
        minutes: now.getMinutes(),
        seconds: now.getSeconds(),
        mseconds: now.getMilliseconds()
    }
    return {
        hours:
            (time.hours / 12) * 360 +
            (time.minutes / 60 / 12) * 360 +
            (time.seconds / 60 / 60 / 12) * 360 +
            (time.mseconds / 1000 / 60 / 60 / 12) * 360,
        minutes: (time.minutes / 60) * 360 + (time.seconds / 60 / 60) * 360 + (time.mseconds / 1000 / 60 / 60) * 360,
        seconds: (time.seconds / 60) * 360 + (time.mseconds / 1000 / 60) * 360
    }
}

const setHands = () => {
    const { hours, minutes, seconds } = getSummedTime()
    hoursHand.style.transform = `rotateZ(${hours}deg)`
    minutesHand.style.transform = `rotateZ(${minutes}deg)`
    secondsHand.style.transform = `rotateZ(${seconds}deg)`
    requestAnimationFrame(setHands)
}

requestAnimationFrame(setHands)
