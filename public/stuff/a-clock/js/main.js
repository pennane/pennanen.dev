let h = document.getElementById("hours")
let m = document.getElementById("minutes")
let s = document.getElementById("seconds")
const getSummedTime = () => {
    return new Promise((resolve) => {
        let now = new Date()
        let time = {
            hours: now.getHours() % 12,
            minutes: now.getMinutes(),
            seconds: now.getSeconds(),
            mseconds: now.getMilliseconds()
        }
        return resolve({
            hours: (time.hours / 12 * 360) 
            + (time.minutes / 60 / 12 * 360) 
            + (time.seconds / 60 / 60 / 12 * 360) 
            + (time.mseconds / 1000 / 60 / 60 / 12 * 360),
            minutes: (time.minutes / 60 * 360) 
            + (time.seconds / 60 / 60 * 360) 
            + (time.mseconds / 1000 / 60 / 60 * 360),
            seconds: (time.seconds / 60 * 360) 
            + (time.mseconds / 1000 / 60 * 360),
        })
    })
}
const setHands = () => {
    getSummedTime().then(deg => {
        h.style.transform = `rotateZ(${deg.hours}deg)`
        m.style.transform = `rotateZ(${deg.minutes}deg)`
        s.style.transform = `rotateZ(${deg.seconds}deg)`
    })
    requestAnimationFrame(setHands)
}

requestAnimationFrame(setHands)