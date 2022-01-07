function degToRad(degrees) {
    return degrees * (Math.PI / 180);
}

function isBetweenRange(num, min, max) {
    if (num >= min && num <= max) {
        return true
    }
    return false
}

function randomBetweenRange(min, max) {
    return Math.floor(Math.random() * (max - min + 1)) + min
}

function randomFromArray(arr) {
    return arr[Math.floor(Math.random() * arr.length)];
}

function numberToString(num) {
    let list = ["nolla", "yksi", "kaksi", "kolme", "neljä", "viisi", "kuusi", "seitsemän", "kahdeksan", "yhdeksän"]
    return list[num]
}

function longNumberToString(num) {
    let list = String(num).split('')
    return list.map((val) => numberToString(val)).join(" ")
}

function pad_with_zeroes(number, length) {
    var my_string = '' + number;
    while (my_string.length < length) {
        my_string = '0' + my_string;
    }
    return my_string;
}

function mapVelocity(velocity) {
    if (velocity > 7) {
        return "erittäin nopea"
    } else if (velocity > 5.3) {
        return "nopea"
    } else if (velocity > 3) {
        return "hidas"
    } else if (velocity > 0) {
        return "erittäin hidas"
    }
}

function mapHeight(height) {
    if (height > 12000) {
        return "erittäin korkealla"
    } else if (height > 4000) {
        return "korkealla"
    } else if (height > 1000) {
        return "matalalla"
    } else if (height > 0) {
        return "erittäin matalalla"
    }
}

function mapDirection(dir) {
    if (isBetweenRange(dir, 337.5, 360) || isBetweenRange(dir, 0, 22.5)) {
        return "pohjoiseen"
    } else if (isBetweenRange(dir, 22.5, 67.5)) {
        return "koilliseen"
    } else if (isBetweenRange(dir, 67.5, 112.5)) {
        return "itään"
    } else if (isBetweenRange(dir, 112.5, 157.5)) {
        return "kaakkoon"
    } else if (isBetweenRange(dir, 157.5, 202.5)) {
        return "etelään"
    } else if (isBetweenRange(dir, 202.5, 247.5)) {
        return "lounaaseen"
    } else if (isBetweenRange(dir, 247.5, 292.5)) {
        return "länteen"
    } else if (isBetweenRange(dir, 292.5, 337.5)) {
        return "luoteeseen"
    }
}

function execute() {


    let squares = [
        {
            letter: "A",
            pronounciation: "alfa"
        },
        {
            letter: "B",
            pronounciation: "pravo"
        },
        {
            letter: "C",
            pronounciation: "tsaali"
        },
        {
            letter: "D",
            pronounciation: "deltta"
        },
        {
            letter: "E",
            pronounciation: "ekko"
        },
        {
            letter: "F",
            pronounciation: "fokstrot"
        },
        {
            letter: "G",
            pronounciation: "golf"
        },
        {
            letter: "H",
            pronounciation: "hotel"
        },
        {
            letter: "J",
            pronounciation: "tsuliet"
        },
        {
            letter: "K",
            pronounciation: "kilo"
        },
        {
            letter: "L",
            pronounciation: "li-ma"
        },
        {
            letter: "M",
            pronounciation: "maik"
        },
        {
            letter: "N",
            pronounciation: "november"
        },
        {
            letter: "P",
            pronounciation: "papa"
        },
        {
            letter: "R",
            pronounciation: "roomeo"
        },
        {
            letter: "S",
            pronounciation: "sierra"
        },
        {
            letter: "T",
            pronounciation: "tango"
        },
        {
            letter: "U",
            pronounciation: "juniform"
        },
        {
            letter: "V",
            pronounciation: "violet"
        },
        {
            letter: "W",
            pronounciation: "viski"
        },
        {
            letter: "X",
            pronounciation: "eksrei"
        },
        {
            letter: "Y",
            pronounciation: "jenkki"
        },
        {
            letter: "Z",
            pronounciation: "tsulu"
        },
    ]

    let aircraftTypes = ["helikopteri", "tiedustelukone", "hävittäjä", "johtokone", "kuljetuskone", "tunnistamaton", "rynnäkkökone"]
    let flyingTactics = ["jatkaa summassa", "pinnassa", "kiertää ilmassa", "etenee", "", "", ""]


    class responsiveCanvas {
        constructor(target, options = { resize: true, fromOffset: true, translateToCenter: true }) {
            this.canvas = target
            this.coordinates = [0, 0]
            this.ctx = target.getContext('2d')


            if (options.fromOffset) {
                this.canvas.width = target.offsetWidth
                this.canvas.height = target.offsetHeight
            } else {
                this.canvas.width = options.width | target.width
                this.canvas.height = options.height | target.height
            }

            if (options.mousemove) {
                this.canvas.addEventListener('mousemove', (e) => {
                    this.setCoordinates(e)
                })
            }
            if (options.resize) {
                this.scheduled = null
                window.addEventListener('resize', (e) => {
                    if (!this.scheduled) {
                        setTimeout(() => {
                            this.resize()
                            this.scheduled = null
                        }, 500)
                    }
                    this.scheduled = true
                })
            }
            if (options.translateToCenter) {
                let transX = this.canvas.width * 0.5,
                    transY = this.canvas.height * 0.5;
                this.ctx.translate(transX, transY);
            }

        }
        getMousePos(event) {
            let rect = this.canvas.getBoundingClientRect()
            let scaleX = this.canvas.width / rect.width
            let scaleY = this.canvas.height / rect.height
            let coordinates = [((event.clientX - rect.left) * scaleX), ((event.clientY - rect.top) * scaleY)]
            return coordinates
        }
        setCoordinates(event) {
            this.coordinates = this.getMousePos(event);
        }
        resize() {
            this.canvas.width = this.canvas.offsetWidth
            this.canvas.height = this.canvas.offsetHeight
        }
    }

    class Vector {
        constructor(x, y) {
            this.x = x
            this.y = y
        }
        plus(vector) {
            return new Vector(this.x + vector.x, this.y + vector.y)
        }
        minus(vector) {
            return new Vector(this.x - vector.x, this.y - vector.y)
        }
        distance(vector) {
            return Math.hypot(vector.x - this.x, vector.y - this.y)
        }
        signedDistance(circle) {
            let distance = this.distance(circle.pos)
            return distance - circle.r
        }
        get length() {
            return Math.sqrt(this.x * this.x + this.y * this.y)
        }

    }

    class DirVector extends Vector {
        constructor(angle, len) {
            super(Math.cos(angle) * len, Math.sin(angle) * len)
        }
    }


    class Plane {
        constructor(options) {
            this.aircrafts = []
            this.dimensions = [5000, 5000] // 10 = 1 km, 1000 = 100 km
            this.offset = [5, 11]
        }
        nextTurn() {
            if (Math.random() > 0.985) {
                let aircraft = generateAircraft()
                this.aircrafts.push(aircraft)
            }
            this.aircrafts.forEach(aircraft => aircraft.updateLocation(aircraft.calculateNextLocation()))
        }
        translatePositionToSquare(location) {
            let square = [squares[this.offset[0] + Math.floor(location.x / 1000)], squares[this.offset[1] + Math.floor(location.y / 1000)]]
            return {
                n: square[0],
                e: square[1],
                nd: parseInt(location.y % 1000 / 100) + 1,
                ed: parseInt(location.x % 1000 / 100) + 1
            }
        }

    }

    let currentId = 1;

    class Aircraft {
        constructor(options) {
            this.location = new Vector(options.location.x, options.location.y) || new Vector(0, 0) // in x, y where one step equals to one hundred meters
            this.direction = options.direction || null // in degrees 0-360 where north is 0, east 90, south 180 etc. if left null, stationary
            this.velocity = options.velocity || 50 // in meters per second
            this.height = options.height || 12000  // height in meters
            this.type = options.type || "unknown" // Leader, assault planes, patrols, fighter jets, helicopter, airplane unidentified, unknown, reconnaissance plane, transport plane
            this.amount = options.amount || 1
            this.possibleInstructions = []
            this.id = pad_with_zeroes(currentId, 4)
            currentId++;
            this.firstEncounter = true;
            this.flyingTactic = randomFromArray(flyingTactics)
            this.removed = false
            this.unusable = false
        }

        calculateNextLocation() {
            let dir = new DirVector(degToRad(this.direction), this.velocity)
            let newLocation = this.location.plus(dir)
            return newLocation
        }
        updateLocation(newLocation) {
            if (Math.random() > 0.985 || this.removed) {
                this.removed = true;
                return
            }
            this.location = newLocation
        }
    }


    const plane = new Plane({ aircrafts: [], dimensions: [5000, 5000] })

    function generateAircraft() {
        let aircraft = new Aircraft({
            location: {
                x: randomBetweenRange(-500, plane.dimensions[0] + 100),
                y: randomBetweenRange(-500, plane.dimensions[1] + 100)
            },
            direction: randomBetweenRange(0, 360),
            velocity: randomBetweenRange(1, 9),
            height: randomBetweenRange(100, 16000),
            type: randomFromArray(aircraftTypes)
        })
        return aircraft
    }

    for (let i = 0; i < 8; i++) {
        let aircraft = generateAircraft()
        plane.aircrafts.push(aircraft)
    }

    let canvasElement = document.getElementById("canvas")
    let responsive = new responsiveCanvas(canvasElement, {
        fromOffset: true,
        resize: false,
        translateToCenter: true
    })

    let canvas = responsive.canvas
    let ctx = responsive.ctx

    let planeTurncounter = setInterval(() => {
        plane.nextTurn()
    }, 4000);

    function playVoice() {
        let readableAircraft = randomFromArray(plane.aircrafts)
        if (readableAircraft.removed) {
            return
        }
        let aircraftPosition = plane.translatePositionToSquare(readableAircraft.location)
        let message = new SpeechSynthesisUtterance();
        message.lang = "fi"
        message.rate = 0.5;
        message.text = `
        ${readableAircraft.firstEncounter ? 'uusi maali' : 'maali'}! . ! . ! . ! .
        ${longNumberToString(readableAircraft.id)}! . ! .
        ${aircraftPosition.e.pronounciation}! . ! . ! . ! . ! .
        ${aircraftPosition.n.pronounciation}! . ! . ! . ! . ! .
        ${numberToString(aircraftPosition.ed)}! . ! . ! . ! . ! .
        ${numberToString(aircraftPosition.nd)}! . ! . ! . ! . ! .
        ${mapDirection(readableAircraft.direction)}! . ! . ! . ! .
        ${mapVelocity(readableAircraft.velocity)}! . ! . ! . ! .
        ${mapHeight(readableAircraft.height)}! . ! . ! . ! . ! .
        ${readableAircraft.type}! . ! . ! . ! . ! .
        ${readableAircraft.flyingTactic}
    `
        if (readableAircraft.removed) {
            readableAircraft.unusable = true
            message.text = `
        ${readableAircraft.firstEncounter ? 'uusi maali' : 'maali'}
        ${longNumberToString(readableAircraft.id)}
        poistettu
        `
            let index = plane.aircrafts.indexOf(readableAircraft)
            plane.aircrafts.splice(index, 1);

        }
        readableAircraft.firstEncounter = false;
        window.speechSynthesis.speak(message);
    }


    let voicecounter = setInterval(() => {
        playVoice()
    }, 40000);

    setTimeout(() => playVoice(), randomBetweenRange(3000, 15000))

    function coordinateLines() {
        ctx.moveTo(-ctx.canvas.width / 2, 0);
        ctx.lineTo(ctx.canvas.width, 0);
        ctx.stroke();
        ctx.beginPath();
        ctx.moveTo(0, -ctx.canvas.height / 2);
        ctx.lineTo(0, ctx.canvas.height / 2);
        ctx.stroke();
        let divided = plane.dimensions[0] / 100
        let squareAmountX = plane.dimensions[0] / 1000
        let squareAmountY = plane.dimensions[1] / 1000
        for (let i = 0; i < squareAmountX; i++) {
            for (let j = 0; j < squareAmountY; j++) {
                ctx.font = "70px Arial";
                ctx.strokeStyle = 'black'
                ctx.lineWidth = 3;
                let rectWidth = canvas.width / squareAmountX
                let rectHeight = canvas.height / squareAmountY
                let startX = -canvas.width / 2 + i * rectWidth;
                let startY = -canvas.height / 2 + j * rectHeight;
                ctx.beginPath()
                ctx.fillStyle = 'rgba(0,0,0,0.5)'
                ctx.fillText(`${squares[j + plane.offset[1]].letter}${squares[i + plane.offset[0]].letter}`, (startX + rectWidth / 2) - 40, (startY + rectHeight / 2) + 20);
                ctx.rect(startX, startY, rectWidth, rectHeight)
                ctx.stroke();
                ctx.fillStyle = 'rgba(0,0,0,1)'
                ctx.lineWidth = 1;
                ctx.strokeStyle = 'black'
            }
        }
        for (let i = 0; i < divided; i++) {
            ctx.beginPath();
            ctx.moveTo(-canvas.width / 2 + (canvas.width / divided * i), -canvas.height / 2);
            ctx.lineTo(-canvas.width / 2 + (canvas.width / divided * i), canvas.height / 2);
            ctx.stroke();
            ctx.beginPath();
            ctx.moveTo(-canvas.width / 2, -canvas.height / 2 + (canvas.height / divided * i));
            ctx.lineTo(canvas.width / 2, -canvas.height / 2 + (canvas.height / divided * i));
            ctx.stroke();
        }

    }

    const animate = () => {
        ctx.fillStyle = '#282636'
        ctx.fillRect(0 - ctx.canvas.width / 2, 0 - ctx.canvas.height / 2, ctx.canvas.width, ctx.canvas.height);
        coordinateLines()
        ctx.stroke()
        ctx.fillStyle = 'black'
        plane.aircrafts.forEach(aircraft => {
            if (aircraft.removed) return
            ctx.beginPath();
            let location = aircraft.location;
            let x = (location.x - 0) / (plane.dimensions[0] - 0) * (canvas.width / 2 - -canvas.width / 2) + - canvas.width / 2;
            let y = (location.y - 0) / (plane.dimensions[1] - 0) * (canvas.height / 2 - -canvas.height / 2) + - canvas.height / 2;
            ctx.fillStyle = '#c22a2a'
            ctx.fillRect(x - 3, y - 3, 6, 6);
            ctx.stroke()
            ctx.fillStyle = 'black'
        })

        requestAnimationFrame(animate)
    }
    requestAnimationFrame(animate)

}

let startButton = document.getElementById('start')

startButton.addEventListener('click', () => {
    startButton.parentNode.removeChild(startButton)
    execute()
})