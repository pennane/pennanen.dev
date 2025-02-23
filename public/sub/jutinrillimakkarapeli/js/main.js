let started = false

const input1 = document.getElementById('input1')
const input2 = document.getElementById('input2')
const input3 = document.getElementById('input3')
const gameImage = document.getElementById('gameImage')

const sounds = {
    start: new Audio('sound/alku.mp3'),
    give: new Audio('sound/anna.mp3'),
    accept: new Audio('sound/joo.mp3'),
    deny: new Audio('sound/ei.mp3')
}

function playSound(sound) {
    try {
        sound.play()
    } catch (err) {}
}

window.addEventListener('click', () => {
    if (started) return
    started = true
    document.getElementById('cover').style.display = 'none'
    playSound(sounds.start)
})

const buttons = document.querySelectorAll('.game-inputs a')
for (const button of buttons) {
    button.addEventListener('click', handleGameState)
}

function begin() {
    input1.innerHTML = 'hmmm anna jutille makkara'
    input2.innerHTML = 'joo'
    input3.innerHTML = 'en'

    input1.setAttribute('data-state-id', '')
    input2.setAttribute('data-state-id', 'acceptState')
    input3.setAttribute('data-state-id', 'denyState')

    gameImage.src = 'img/ahm.jpg'
    playSound(sounds.give)
}

function accept() {
    input1.innerHTML = ''
    input2.innerHTML = ''
    input3.innerHTML = ''

    input1.setAttribute('data-state-id', '')
    input2.setAttribute('data-state-id', '')
    input3.setAttribute('data-state-id', '')

    gameImage.src = 'img/happy.jpg'

    playSound(sounds.accept)

    setTimeout(function () {
        end()
    }, 2700)
}

function deny() {
    input1.innerHTML = ''
    input2.innerHTML = ''
    input3.innerHTML = ''

    input1.setAttribute('data-state-id', '')
    input2.setAttribute('data-state-id', '')
    input3.setAttribute('data-state-id', '')

    gameImage.src = 'img/sad.jpg'

    playSound(sounds.deny)

    setTimeout(function () {
        end()
    }, 2700)
}

function end() {
    input1.innerHTML = 'Aloita peli'
    input2.innerHTML = ''
    input3.innerHTML = ''

    input1.setAttribute('data-state-id', 'beginState')
    input2.setAttribute('data-state-id', '')
    input3.setAttribute('data-state-id', '')

    gameImage.src = 'img/fas.jpg'

    playSound(sounds.start)
}

function handleGameState(event) {
    console.log(event)
    const gameState = event.target.dataset.stateId
    switch (gameState) {
        case 'beginState':
            begin()
            break
        case 'acceptState':
            accept()
            break
        case 'denyState':
            deny()
            break
    }
}

if (window.location.hash && window.location.hash == '#easteregg') {
    document.body.style.backgroundImage = "url('img/sad.gif')"
    document.body.style.backgroundRepeat = 'repeat'
    document.body.style.animation = 'huerotation 1s linear infinite'
    document.body.style.WebkitAnimation = 'huerotation 1s linear infinite'
}
