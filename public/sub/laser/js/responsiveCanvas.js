export default class responsiveCanvas {
    constructor(target, options = { resize: true, fromOffset: true }) {
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
        if (options.translate) {
            let transX = this.canvas.width * 0.5
            let transY = this.canvas.height * 0.5
            this.ctx.translate(transX, transY)
        }
        this.options = options

    }
    getMousePos(event) {
        let rect = this.canvas.getBoundingClientRect()
        let scaleX = this.canvas.width / rect.width
        let scaleY = this.canvas.height / rect.height
        let coordinates = [((event.clientX - rect.left) * scaleX), ((event.clientY - rect.top) * scaleY)]
        if (this.options.translate) {
            coordinates[0] -= this.canvas.width / 2
            coordinates[1] -= this.canvas.height / 2
        }
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