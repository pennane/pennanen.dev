class Storage {
    constructor() {
        this.storage = window.localStorage
    }
    set(item, value) {
        return this.storage.setItem(item, value)
    }
    get(item) {
        return this.storage.getItem(item)
    }
    has(item) {
        return !!this.storage.getItem(item)
    }
    remove(item) {
        return this.storage.removeItem(item)
    }

    keys() {
        return Object.keys(this.storage)
    }

    clear() {
        return this.storage.clear()
    }

    isEmpty() {
        return this.storage.length > 0
    }
}

let storage = new Storage()

let gameData

if (storage.has('game-data')) {
    gameData = JSON.parse(localStorage.getItem('game-data'))
} else {
    gameData = {
        scores: {
            fishflockscore: 0,
            fishhutscore: 0,
            fishmagicscore: 0,
            fishmanscore: 0,
            fishvesselscore: 0
        },
        utilities: {
            autoclicker: 0,
            fishflock: {
                price: 50000
            },
            fishhut: {
                price: 500
            },
            fishmagic: {
                price: 1000000
            },
            fishman: {
                price: 20
            },
            fishvessel: {
                price: 2000
            },
            amountOfFish: 0,
            nE: 0,
            totaltotalyield: 0,
            totalyield: 0
        },
        lucrativity: {
            fishflockyield: 100,
            fishhutyield: 4,
            fishmagicyield: 10000,
            fishmanyield: 0.2,
            fishvesselyield: 40,
            fishyield: 1
        },
        upgrades: {
            autoFisher: {
                price: 1000,
                purchased: false
            },
            all2x: {
                price: 10000,
                purchased: false,
                multiplier: 2
            },
            fishermen2x: {
                price: 20000,
                purchased: false,
                multiplier: 2
            },
            fishinghut2x: {
                price: 50000,
                purchased: false,
                multiplier: 2
            },
            fishvessel2x: {
                price: 100000,
                purchased: false,
                multiplier: 2
            },
            fishflock2x: {
                price: 500000,
                purchased: false,
                multiplier: 2
            },
            all4x: {
                price: 10000,
                purchased: false,
                multiplier: 4
            },
            fishermagic2x: {
                price: 1000000,
                purchased: false,
                multiplier: 2
            }
        },
        isFirstTime: true
    }
}

var firstTime = storage.get(gameData.isFirstTime)

let resetButton = document.getElementById('resetbtn')
let progressButton = document.getElementById('prognappi')
let statsButton = document.getElementById('statsnappi')
let fishingButton = document.getElementById('fishingbtn')

resetButton.addEventListener('click', () => {
    if (confirm('Are you sure?')) {
        abortTimer()
        ckremove()
        Cookies.remove('firsttime')
        setTimeout(function () {
            location.reload()
        }, 3000)
    }
})

progressButton.addEventListener('click', () => {
    let progress = document.getElementById('progress')
    let stats = document.getElementById('stats')

    progress.style.opacity = '0'
    stats.style.opacity = '1'

    progress.style.zIndex = '-1'
    stats.style.zIndex = 'auto'

    progressButton.classList.add('active')
    statsButton.classList.remove('active')
})

statsButton.addEventListener('click', () => {
    let progress = document.getElementById('progress')
    let stats = document.getElementById('stats')

    progress.style.opacity = '1'
    stats.style.opacity = '0'

    progress.style.zIndex = 'auto'
    stats.style.zIndex = '-1'

    progressButton.classList.remove('active')
    statsButton.classList.add('active')
})

let clicked = null
fishingButton.addEventListener('click', () => {
    clicked = Date.now()
    const oldClicked = clicked

    document.body.classList.add('fishing')
    addscore('fish')
    floatingnumbers()
    setTimeout(function () {
        if (oldClicked === clicked) {
            document.body.classList.remove('fishing')
        }
    }, 2000)
})

if (firstTime) {
    storage.set(firstTime, false)
    document.querySelector('.attention').style.opacity = '1'
    setTimeout(function () {
        document.querySelector('.attention').style.opacity = '0'
    }, 10000)
}

function floatingnumbers() {
    gameData.utilities.nE++
    let floatingNumber = document.createElement('span')

    floatingNumber.className = 'floatingnumber'
    floatingNumber.id = 'num' + gameData.utilities.nE
    let fishPerClick =
        gameData.lucrativity.fishyield *
        (gameData.upgrades.all2x.purchased ? gameData.upgrades.all2x.multiplier : 1) *
        (gameData.upgrades.all4x.purchased ? gameData.upgrades.all4x.multiplier : 1)
    floatingNumber.textContent = parseInt(fishPerClick)
    document.querySelector('#clickbtnholder').appendChild(floatingNumber)
    setTimeout(() => {
        floatingNumber.parentElement.removeChild(floatingNumber)
    }, 1500)
}

function purchase(val) {
    switch (val) {
        case 'fishman':
            if (gameData.utilities.amountOfFish >= gameData.utilities.fishman.price) {
                gameData.utilities.amountOfFish -= gameData.utilities.fishman.price
                gameData.utilities.fishman.price = parseInt(gameData.utilities.fishman.price * 1.2)
                gameData.scores.fishmanscore++
                updatestats()
            }
            break

        case 'fishhut':
            if (gameData.utilities.amountOfFish >= gameData.utilities.fishhut.price) {
                gameData.utilities.amountOfFish -= gameData.utilities.fishhut.price
                gameData.utilities.fishhut.price = parseInt(gameData.utilities.fishhut.price * 1.2)
                gameData.scores.fishhutscore = isNaN(gameData.scores.fishhutscore)
                    ? 1
                    : gameData.scores.fishhutscore + 1
                updatestats()
            }
            break

        case 'fishvessel':
            if (gameData.utilities.amountOfFish >= gameData.utilities.fishvessel.price) {
                gameData.utilities.amountOfFish -= gameData.utilities.fishvessel.price
                gameData.utilities.fishvessel.price = parseInt(gameData.utilities.fishvessel.price * 1.2)
                gameData.scores.fishvesselscore = isNaN(gameData.scores.fishvesselscore)
                    ? 1
                    : gameData.scores.fishvesselscore + 1
                updatestats()
            }
            break

        case 'fishflock':
            if (gameData.utilities.amountOfFish >= gameData.utilities.fishflock.price) {
                gameData.utilities.amountOfFish -= gameData.utilities.fishflock.price
                gameData.utilities.fishflock.price = parseInt(gameData.utilities.fishflock.price * 1.2)
                gameData.scores.fishflockscore = isNaN(gameData.scores.fishflockscore)
                    ? 1
                    : gameData.scores.fishflockscore + 1
                updatestats()
            }
            break
        case 'fishmagic':
            if (gameData.utilities.amountOfFish >= gameData.utilities.fishmagic.price) {
                gameData.utilities.amountOfFish -= gameData.utilities.fishmagic.price
                gameData.utilities.fishmagic.price = parseInt(gameData.utilities.fishmagic.price * 1.2)
                gameData.scores.fishmagicscore = isNaN(gameData.scores.fishmagicscore)
                    ? 1
                    : gameData.scores.fishmagicscore + 1
                updatestats()
            }
            break
        case 'upg1':
            if (
                gameData.utilities.amountOfFish >= gameData.upgrades.all2x.price &&
                !gameData.upgrades.all2x.purchased
            ) {
                gameData.utilities.amountOfFish -= gameData.upgrades.all2x.price
                updatestats()
                gameData.upgrades.all2x.purchased = true
            }
            break
        case 'upg2':
            if (
                gameData.utilities.amountOfFish >= gameData.upgrades.fishermen2x.price &&
                !gameData.upgrades.fishermen2x.purchased
            ) {
                gameData.utilities.amountOfFish -= gameData.upgrades.fishermen2x.price
                gameData.lucrativity.fishmanyield *= 2
                updatestats()
                gameData.upgrades.fishermen2x.purchased = true
            }
            break
        case 'upg3':
            if (
                gameData.utilities.amountOfFish >= gameData.upgrades.fishinghut2x.price &&
                !gameData.upgrades.fishinghut2x.purchased
            ) {
                gameData.utilities.amountOfFish -= gameData.upgrades.fishinghut2x.price
                gameData.lucrativity.fishhutyield *= 2
                updatestats()
                gameData.upgrades.fishinghut2x.purchased = true
            }
            break
        case 'upg4':
            if (
                gameData.utilities.amountOfFish >= gameData.upgrades.fishvessel2x.price &&
                !gameData.upgrades.fishvessel2x.purchased
            ) {
                gameData.utilities.amountOfFish -= gameData.upgrades.fishvessel2x.price
                gameData.lucrativity.fishvesselyield *= 2
                updatestats()
                gameData.upgrades.fishvessel2x.purchased = true
            }
            break
        case 'upg5':
            if (
                gameData.utilities.amountOfFish >= gameData.upgrades.fishflock2x.price &&
                !gameData.upgrades.fishflock2x.purchased
            ) {
                gameData.utilities.amountOfFish -= gameData.upgrades.fishflock2x.price
                gameData.lucrativity.fishflockyield *= 2
                updatestats()
                gameData.upgrades.fishflock2x.purchased = true
            }
            break
        case 'upg6':
            if (
                gameData.utilities.amountOfFish >= gameData.upgrades.fishermagic2x.price &&
                !gameData.upgrades.fishermagic2x.purchased
            ) {
                gameData.utilities.amountOfFish -= gameData.upgrades.fishermagic2x.price
                gameData.lucrativity.fishmagicyield *= 2
                updatestats()
                gameData.upgrades.fishermagic2x.purchased = true
            }
            break
        case 'upg7':
            if (
                gameData.utilities.amountOfFish >= gameData.upgrades.all4x.price &&
                !gameData.upgrades.all4x.purchased
            ) {
                gameData.utilities.amountOfFish -= gameData.upgrades.all4x.price
                updatestats()
                gameData.upgrades.all4x.purchased = true
            }
            break
        case 'upg8':
            if (
                gameData.utilities.amountOfFish >= gameData.upgrades.autoFisher.price &&
                !gameData.upgrades.autoFisher.purchased
            ) {
                gameData.utilities.amountOfFish -= gameData.upgrades.autoFisher.price
                gameData.upgrades.autoFisher.purchased = true
            }
            break
    }
}

function addscore(val) {
    switch (val) {
        case 'fish':
            let fishAdd = gameData.lucrativity.fishyield
            if (gameData.upgrades.all2x.purchased) {
                fishAdd *= gameData.upgrades.all2x.multiplier
            }

            if (gameData.upgrades.all4x.purchased) {
                fishAdd *= gameData.upgrades.all4x.multiplier
            }

            gameData.utilities.amountOfFish += parseInt(fishAdd)
            document.querySelector('.fishstat').textContent = gameData.utilities.amountOfFish
            break

        case 'fishman':
            gameData.scores.fishmanscore++
            document.querySelector('#fishmanstat').textContent = gameData.scores.fishmanscore
            break

        case 'fishhut':
            gameData.scores.fishhutscore++
            document.querySelector('#fishhutstat').textContent = gameData.scores.fishhutscore
            break

        case 'fishvessel':
            gameData.scores.fishvesselscore++
            document.querySelector('#fishvesselstat').textContent = gameData.scores.fishvesselscore
            break

        case 'fishflock':
            gameData.scores.fishflockscore++
            document.querySelector('#fishflockstat').textContent = gameData.scores.fishflockscore
            break

        case 'fishmagic':
            gameData.scores.fishmagicscore++
            document.querySelector('#fishmagicstat').textContent = gameData.scores.fishmagicscore
            break
    }
}

var tid = setTimeout(fishpersec, 1000)

function fishpersec() {
    totalyield =
        gameData.lucrativity.fishmanyield * gameData.scores.fishmanscore +
        gameData.lucrativity.fishhutyield * gameData.scores.fishhutscore +
        gameData.lucrativity.fishvesselyield * gameData.scores.fishvesselscore +
        gameData.lucrativity.fishflockyield * gameData.scores.fishflockscore +
        gameData.lucrativity.fishmagicyield * gameData.scores.fishmagicscore

    totaltotalyield = totalyield

    if (gameData.upgrades.all2x.purchased) {
        totaltotalyield *= gameData.upgrades.all2x.multiplier
    }

    if (gameData.upgrades.all4x.purchased) {
        totaltotalyield *= gameData.upgrades.all4x.multiplier
    }

    totaltotalyield = Math.round(totaltotalyield * 100) / 100

    gameData.utilities.amountOfFish += totaltotalyield

    gameData.utilities.amountOfFish = Math.round(gameData.utilities.amountOfFish * 100) / 100

    let fishstat = document.querySelector('.fishstat')

    if (gameData.utilities.amountOfFish > 10000000) {
        fishstat.textContent = gameData.utilities.amountOfFish.toExponential()
    } else {
        fishstat.textContent = gameData.utilities.amountOfFish
    }

    let fishpersecstat = document.querySelector('#fishpersecstat')
    if (totaltotalyield > 1000000) {
        fishpersecstat.textContent = totaltotalyield.toExponential()
    } else {
        fishpersecstat.textContent = totaltotalyield
    }

    updatestats()

    if (!document.getElementById('autoclicker') && gameData.upgrades.autoFisher.purchased) {
        let target = document.querySelector('#utilitiescontent')
        let autoclickerElement = document.createElement('div')
        autoclickerElement.id = 'autoclicker'
        autoclickerElement.className = 'utility'
        let s = document.createElement('span')
        s.textContent = 'Autoclicker'
        let b = document.createElement('button')
        b.textContent = 'on / off'
        b.id = 'autoOnOff'

        autoclickerElement.appendChild(s)
        autoclickerElement.appendChild(b)

        target.appendChild(autoclickerElement)

        document.getElementById('autoOnOff').addEventListener('click', () => {
            if (utilities.autoclicker === 0) {
                utilities.autoclicker = 1
            } else {
                utilities.autoclicker = 0
            }
            autoclickerloop()
        })
    }

    if (document.getElementById('autoclicker') && utilities.autoclicker === 1) {
        let autoClickerStatus = document.getElementById('autoOnOff')
        autoClickerStatus.textContent = 'on'
        autoClickerStatus.style.filter = ''
    }
    if (document.getElementById('autoclicker') && utilities.autoclicker === 0) {
        let autoClickerStatus = document.getElementById('autoOnOff')
        autoClickerStatus.textContent = 'off'
        autoClickerStatus.style.filter = 'brightness: 40'
    }

    tid = setTimeout(fishpersec, 1000)
}

;[...document.querySelectorAll('.purchasingButton')].forEach((button) => {
    button.addEventListener('click', () => {
        if (button.value) {
            purchase(button.value)
        } else {
            purchase(button.id)
        }
    })
})

function autoclickerloop() {
    setTimeout(function () {
        fishingButton.click()
        if (utilities.autoclicker === 1) {
            autoclickerloop()
        }
    }, 250)
}

function cktimer() {
    storage.set('game-data', JSON.stringify(gameData))
    cookietimer = setTimeout(cktimer, 8000)
}

let cookietimer = setTimeout(() => cktimer(), 8000)

function abortTimer() {
    clearTimeout(tid)
    clearTimeout(cookietimer)
}

function updatestats() {
    /* Scores */
    ;[...document.querySelectorAll('.fishstat')].forEach((el) => (el.textContent = gameData.utilities.amountOfFish))
    document.querySelector('#fishmanstat').textContent = gameData.scores.fishmanscore
    document.querySelector('#fishhutstat').textContent = gameData.scores.fishhutscore
    document.querySelector('#fishvesselstat').textContent = gameData.scores.fishvesselscore
    document.querySelector('#fishflockstat').textContent = gameData.scores.fishflockscore
    document.querySelector('#fishmagicstat').textContent = gameData.scores.fishmagicscore

    /* Prices */

    document.querySelector('#fishmanprice').textContent = gameData.utilities.fishman.price
    document.querySelector('#fishhutprice').textContent = gameData.utilities.fishhut.price
    document.querySelector('#fishvesselprice').textContent = gameData.utilities.fishvessel.price
    document.querySelector('#fishflockprice').textContent = gameData.utilities.fishflock.price
    document.querySelector('#fishmagicprice').textContent = gameData.utilities.fishmagic.price

    /* Per secs */

    document.querySelector('#fishmanpersec').textContent =
        gameData.lucrativity.fishmanyield * gameData.scores.fishmanscore
    document.querySelector('#fishhutpersec').textContent =
        gameData.lucrativity.fishhutyield * gameData.scores.fishhutscore
    document.querySelector('#fishvesselpersec').textContent =
        gameData.lucrativity.fishvesselyield * gameData.scores.fishvesselscore
    document.querySelector('#fishflockpersec').textContent =
        gameData.lucrativity.fishflockyield * gameData.scores.fishflockscore
    document.querySelector('#fishmagicpersec').textContent =
        gameData.lucrativity.fishmagicyield * gameData.scores.fishmagicscore
}
