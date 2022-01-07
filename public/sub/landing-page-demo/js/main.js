let targetEl = document.querySelector(".hero")

const rnd = (min, max) => Math.floor(Math.random() * (max - min)) + min;

const getAmount = (w) => Math.floor((w / 100) * 5)

let currentAmount = getAmount(window.innerWidth)

function createBubbles(amount) {
	amount = amount > 100 ? 100 : amount
	let old = document.querySelector(".heroanimator");
	let bubbleParent;
	if (old) {
		old.innerHTML = "";
		bubbleParent = old;
	} else {
		bubbleParent = document.createElement('div')
		bubbleParent.className = "heroanimator"
	}
	let duration = 30;
	for (let i = 0; i < amount; i++) {
		let el = document.createElement('div')
		el.style.top = `${rnd(5, 95)}%`;
		el.style.left = `${rnd(5, 95)}%`;
		el.className = "bubble";
		el.style.animationDelay = `${-(i * duration * (Math.random() * 2))}s`;
		el.style.animationDuration = `${duration}s`
		bubbleParent.appendChild(el)
	}
	return bubbleParent
}

function initBubbles() {
	let amount = getAmount(window.innerWidth)
	targetEl.appendChild(createBubbles(amount))
}

let timeout;
window.addEventListener("resize", event => {
	clearTimeout(timeout)
	timeout = setTimeout(() => {
		let changedAmount = getAmount(window.innerWidth);
		if (Math.abs(currentAmount - changedAmount ) > 5 ) {
			initBubbles()
			currentAmount = changedAmount;
		}
		
	}, 400)
})

initBubbles()

