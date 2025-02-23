const rnd = (min, max) => (Math.random() * (max - min)) + min;

class responsiveCanvas {
    constructor(target, options = {
        resize: true,
        fromOffset: true
    }) {
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

class Circle {
    constructor(x, y, r) {
        this.pos = new Vector(x, y)
        this.r = r
    }

    show() {
        ctx.beginPath()
        ctx.arc(this.pos.x, this.pos.y, this.r, 0, 2 * Math.PI);
        ctx.stroke();
        ctx.closePath()
    }
}

class Ray {
    constructor(x, y, angle) {
        this.pos = new Vector(x, y)
        this.angle = angle
    }

    march(items) {
        let i = 0
        let closest = this.closest(items, this.pos)
        let dist = this.pos.signedDistance(closest)
        let current = new Vector(this.pos.x, this.pos.y)
        while (i < 100) {
            if (dist < 1 || dist > window.innerWidth / 2.1) break
            let direction = new DirVector(this.angle, dist)

            ctx.beginPath()
            ctx.moveTo(current.x, current.y)

            let to = current.plus(direction)
            ctx.strokeStyle = "#1f00ff"
            ctx.lineTo(to.x, to.y)
            ctx.stroke()
            ctx.closePath()
            ctx.beginPath()
            ctx.arc(current.x, current.y, dist, 0, 2 * Math.PI);
            ctx.strokeStyle = "#ff002f"
            ctx.stroke()
            ctx.strokeStyle = "#000000"
            ctx.closePath()

            current = current.plus(direction)
            closest = this.closest(items, current)
            dist = current.signedDistance(closest)
            i++
        }


    }
    closest(items, point) {
        return items.reduce((a, b) => point.signedDistance(a) < point.signedDistance(b) ? a : b)
    }
}

let el = document.getElementById("cnvs")
let responsive = new responsiveCanvas(el, {
    fromOffset: true,
    resize: false,
    mousemove: true,
    width: 600,
    height: 600
})

let canvas = responsive.canvas
let ctx = responsive.ctx
ctx.lineWidth = 2
let amount = parseInt((window.innerWidth * window.innerHeight) / 245760) + 5

let ray;
let circles = [];
let animate;

function init() {
    ray = new Ray(canvas.width / 2, canvas.height / 2, 0)
    circles = []
    for (let i = 0; i < amount; i++) {
        let circle = new Circle(rnd(0, canvas.width), rnd(0, canvas.height), rnd(20, 100))
        while (ray.pos.signedDistance(circle) < 30) {
            circle = new Circle(rnd(0, canvas.width), rnd(0, canvas.height), rnd(20, 100))
        }
        circles.push(circle)
    }
    

    animate = () => {
        ctx.clearRect(0, 0, ctx.canvas.width, ctx.canvas.height);
        circles.forEach((item) => item.show())
        ray.angle += 0.003
        ray.march(circles)
        requestAnimationFrame(animate)
    }
    
    requestAnimationFrame(animate)
}

function reset() {
    ray = new Ray(canvas.width / 2, canvas.height / 2, 0)
    circles = []
    for (let i = 0; i < amount; i++) {
        let circle = new Circle(rnd(0, canvas.width), rnd(0, canvas.height), rnd(20, 100))
        while (ray.pos.signedDistance(circle) < 30) {
            circle = new Circle(rnd(0, canvas.width), rnd(0, canvas.height), rnd(20, 100))
        }
        circles.push(circle)
    }
    ctx.lineWidth = 2
}

let timeout;
window.addEventListener('resize', (e) => {
    clearTimeout(timeout);
    timeout = setTimeout(() => {
        responsive.resize()
        reset()
        
    }, 50);
  });
init()