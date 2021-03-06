export function clampBetween(input, min, max) {
    return input < min ? min : input > max ? max : input
}

export function mapBetween(current, in_min, in_max, out_min, out_max) {
    const mapped = ((current - in_min) * (out_max - out_min)) / (in_max - in_min) + out_min
    return clampBetween(mapped, out_min, out_max)
}

export function wait(time) {
    return new Promise((resolve) => setTimeout(() => resolve(), time))
}
