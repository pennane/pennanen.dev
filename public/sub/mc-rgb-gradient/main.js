import generateGradient from './generateGradient.js'

let nickNameElement = document.getElementById('nick-name')

let color1Element = document.getElementById('first-color')
let color2Element = document.getElementById('second-color')
let stepsElement = document.getElementById('steps-amount')

let previewOutputElement = document.getElementById('preview')
let nicknameOutputElement = document.getElementById('nickname-output')
let warningOutputElement = document.getElementById('warning')

function generateGradientName(nickName, colorArray) {
    let chunkLength = Math.ceil(nickName.length / colorArray.length)
    let splitNickname = nickName.match(new RegExp('.{1,' + chunkLength + '}', 'g'));
    let completeNickname = "";
    splitNickname.forEach((chunk, i) => {
        completeNickname += `&${colorArray[i]}${chunk}`
    })
    return completeNickname;
}

function parsedNicknameElement(nickName) {
    let output = document.createElement('span')
    let splitKey = "yE9sYcWcZFPo:=kn7,t|k$rP(cDvqR"
    let colorChunks = nickName.replace(/&#/gi, splitKey+"&#").split(splitKey);
    colorChunks.shift()
    colorChunks.forEach((chunk, i) => {
        let chunkEl = document.createElement('span')
        let colorChunk = chunk.substr(0, 8)
        let nameChunk = chunk.replace(colorChunk, '')
        let color = colorChunk.substring(1)
        chunkEl.style.color = color;
        chunkEl.textContent = nameChunk;
        output.appendChild(chunkEl)
    })

    return output;
}

function update() {


    let gradientColors = generateGradient(color1Element, color2Element, stepsElement)

    let gradientNickname = generateGradientName(nickNameElement.value, gradientColors)

    let parsedNicname = parsedNicknameElement(gradientNickname)

    while (previewOutputElement.firstChild) {
        previewOutputElement.removeChild(previewOutputElement.lastChild);
    }

    if (gradientNickname.length > 128) {
        warningOutputElement.textContent = "YOUR NICKNAME IS TOO LONG TO BE USED!"
    } else {
        warningOutputElement.textContent = ""
    }

    previewOutputElement.appendChild(parsedNicname)
    nicknameOutputElement.textContent = gradientNickname
}

let timeout;
document.addEventListener('change', () => {
    clearTimeout(timeout);
    timeout = setTimeout(() => update(), 500);

})

document.addEventListener('input', () => {
    clearTimeout(timeout);
    timeout = setTimeout(() => update(), 500);
})

update()
