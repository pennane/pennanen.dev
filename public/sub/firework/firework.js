let canvas = document.getElementById('canvas')
let ctx = canvas.getContext('2d')
let amountOfRockets = 4
let soundsCache = {
	fly: new Audio('fly.mp3'),
	explosion: new Audio('explosion.mp3')
}

const randInt = (min, max) => Math.floor(Math.random() * (max - min + 1) + min)

class Firework {
	constructor() {
		this.reset()
	}

	reset() {
		this.size = randInt(4,5),
			this.x = randInt(20, canvas.width - 20),
			this.y = randInt(1000, 5) + canvas.height,
			this.speed = randInt(7, 11),
			this.state = 'fly',
			this.sparks = [],
			this.trailAge = 0,
			this.addSparks(randInt(5,50)),
			this.colour = this.getColour(),
			this.age = 0
		/*this.sound = {}
		this.sound.fly = new Audio('fly.mp3')
		this.sound.fly.volume = 0.2
		this.sound.explosion = new Audio('explosion.mp3')
		this.sound.explosion.volume = 0.2
		this.sound.fly.play()*/
	}

	getColour() {
		let colours = ['255, 255, 0', '252, 252, 252', '53, 13, 232', '255,165,0', '12, 214, 19', '196, 48, 11', '0,255,0', '34, 118, 244']
		return colours[Math.floor(Math.random() * colours.length)]
	}

	addSparks(amount) {
		this.sparks = []
		for (let i = 0; i < amount; i++) {
			let spark = {
				vx: Math.random() * 2 + 0.1,
				vy: Math.random() * 2 + 0.1,
				weight: Math.random() * .3 + .03
			}
			if (Math.random() > .5) spark.vx = -spark.vx;
			if (Math.random() > .5) spark.vy = -spark.vy;
			this.sparks.push(spark)
		}
	}

	stateCheck() {
		if (((Math.random() < .015 && this.y < 300) || this.y < 175) && this.state !== 'explode') {
			this.state = 'explode'
			/*this.sound.fly.pause()
			this.sound.explosion.play()*/
		}

		if (this.state === 'fly') this.fly()
		else if (this.state === 'explode') this.explode()
	}

	fly() {
		ctx.beginPath()
		ctx.rect(this.x, this.y, this.size, this.size);
		ctx.fillStyle = `rgba(${this.colour},1)`
		ctx.fill();
		this.y -= this.speed
	}

	explode() {
		if (this.trailAge >= 50) {
			return this.reset()
		};
		this.sparks.forEach(spark => {
			let trailAge = this.trailAge
			let sx = this.x + spark.vx * trailAge
			let sy = this.y + spark.vy * trailAge + (Math.pow(spark.weight, 3) * Math.pow(trailAge, 2))
			let sparkOpacity = (50 - trailAge) / 20
			ctx.beginPath()
			ctx.fillStyle = `rgba(${this.colour},${sparkOpacity})`
			ctx.rect(sx, sy, this.size, this.size)
			ctx.fill()

		})
		this.trailAge++
	}

}

let fireworks = []

for (let i = 0; i < amountOfRockets; i++) {
	let firework = new Firework({})
	fireworks.push(firework)
}

let animateFirework = () => {
	ctx.clearRect(0, 0, canvas.width, canvas.height)
	fireworks.forEach(firework => {
		firework.stateCheck()
	})
	window.requestAnimationFrame(animateFirework)
}

window.requestAnimationFrame(animateFirework)