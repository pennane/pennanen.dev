
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
        while (i < 100 ) {
            if (dist < 1 || dist > window.innerWidth/3) break
            let direction = new DirVector(this.angle, dist)

            ctx.beginPath()
            ctx.moveTo(current.x,current.y)

            let to = current.plus(direction)

            ctx.lineTo(to.x, to.y)
            ctx.stroke()
            ctx.closePath()
            ctx.beginPath()
            ctx.arc(current.x, current.y, dist, 0, 2 * Math.PI);
            ctx.strokeStyle = "#0f00ff"
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