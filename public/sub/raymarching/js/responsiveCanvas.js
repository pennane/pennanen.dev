export default class responsiveCanvas {
    constructor(target, options = {resize: true, fromOffset:true}) {
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