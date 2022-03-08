let started = false

let sounds = {
    start: new Audio('sound/alku.mp3'),
    give: new Audio('sound/anna.mp3'),
    yes: new Audio('sound/joo.mp3'),
    no: new Audio('sound/ei.mp3')
}

const play = (s) => {
    try {
        s.play()
    } catch (err) {}
}

window.addEventListener('click', () => {
    if (started) return
    started = true
    document.getElementById('cover').style.display = 'none'
    play(sounds.start)
})

let buttons = document.querySelectorAll('.game')
buttons.forEach((button) => button.addEventListener('click', game))

function game(event) {
    let val = event.target.id
    if (val == 'aloitus') {
        document.getElementById('aloitus').innerHTML = 'hmmm anna jutille makkara'
        document.getElementById('aloitus').id = 'placeholder1'
        document.getElementById('kuvituskuva').src = 'img/ahm.jpg'

        document.getElementById('aloitus2').innerHTML = 'joo'
        document.getElementById('aloitus2').id = 'joo'

        document.getElementById('aloitus3').innerHTML = 'en'
        document.getElementById('aloitus3').id = 'ei'

        play(sounds.give)
    } else if (val == 'joo') {
        document.getElementById('joo').innerHTML = ''
        document.getElementById('ei').innerHTML = ''
        document.getElementById('placeholder1').innerHTML = ''

        document.getElementById('joo').id = 'placeholder2'
        document.getElementById('ei').id = 'placeholder3'
        document.getElementById('placeholder1').id = 'loppuish'

        document.getElementById('kuvituskuva').src = 'img/happy.jpg'
        play(sounds.yes)
        setTimeout(function () {
            jatka()
        }, 2700)
    } else if (val == 'ei') {
        document.getElementById('joo').innerHTML = ''
        document.getElementById('ei').innerHTML = ''
        document.getElementById('placeholder1').innerHTML = ''
        document.getElementById('joo').id = 'placeholder2'
        document.getElementById('ei').id = 'placeholder3'
        document.getElementById('placeholder1').id = 'loppuish'

        document.getElementById('kuvituskuva').src = 'img/sad.jpg'
        play(sounds.no)
        setTimeout(function () {
            jatka()
        }, 2700)
    }

    function jatka() {
        let sound1 = new Audio('sound/alku.mp3')
        play(sounds.start)
        document.getElementById('loppuish').innerHTML = 'Aloita peli'
        document.getElementById('loppuish').id = 'aloitus'
        document.getElementById('kuvituskuva').src = 'img/fas.jpg'
        document.getElementById('placeholder2').id = 'aloitus2'
        document.getElementById('placeholder3').id = 'aloitus3'
    }
}

if (window.location.hash) {
    let hash = window.location.hash
    if (hash == '#easteregg') {
        document.body.style.backgroundImage = "url('img/sad.gif')"
        document.body.style.backgroundRepeat = 'repeat'
        document.body.style.animation = 'huerotation 1s linear infinite'
        document.body.style.WebkitAnimation = 'huerotation 1s linear infinite'
    }
}
