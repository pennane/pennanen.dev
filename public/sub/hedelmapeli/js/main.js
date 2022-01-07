let s1f1;
let s1f2;
let s1f3;
let s2f1;
let s2f2;
let s2f3;
let s3f1;
let s3f2;
let s3f3;

let off = true;
let a3 = 2;
let b3 = 2;
let o3 = 2;
let k2 = 4;
let k3 = 8;
let t2 = 5;
let t3 = 15;
let p1 = 4;
let p2 = 18;
let p3 = 30;

let fruitClasses = ["b", "a", "o", "k", "t", "p"];

let credits = 100;
let bets = [0.1, 0.2, 0.5, 1, 2, 5, 10];
let bet = 1;

let val1;
let val2;
let val3;
let loppusuora;
let voitto;

let spinbutton = document.getElementById("spinbtn");
let betdown = document.getElementById("betdown");
let betup = document.getElementById("betup");
let winspan = document.getElementById("winspan");
let spinvalue = document.getElementById("spinvalue");
let betvalue = document.getElementById("betvalue");

const randomFrom = arr => arr[Math.floor(Math.random() * arr.length)];

const removeClasses = (el, arr) =>
	arr.forEach(clas => {
		el.classList.toggle(clas, false);
	});

let fruits = [];
let i;

for (i = 0; i < 23; i++) {
	fruits.push("b");
}
/* Banaani  */

for (i = 0; i < 23; i++) {
	fruits.push("a");
}
/* Apelsiini  */

for (i = 0; i < 23; i++) {
	fruits.push("o");
}
/* Omena  */

for (i = 0; i < 17; i++) {
	fruits.push("k");
}
/* Kultaharkko  */

for (i = 0; i < 13; i++) {
	fruits.push("t");
}
/* Timantti  */

for (i = 0; i < 4; i++) {
	fruits.push("p");
}
/* Peigom  */

window.addEventListener("load", () => {
	update();

	s1f1 = randomFrom(fruits);
	s1f2 = randomFrom(fruits);
	s1f3 = randomFrom(fruits);
	s2f1 = randomFrom(fruits);
	s2f2 = randomFrom(fruits);
	s2f3 = randomFrom(fruits);
	s3f1 = randomFrom(fruits);
	s3f2 = randomFrom(fruits);
	s3f3 = randomFrom(fruits);

	document.querySelector("#s1f1 .fruit-image").classList.toggle(s1f1, true);
	document.querySelector("#s2f1 .fruit-image").classList.toggle(s2f1, true);
	document.querySelector("#s3f1 .fruit-image").classList.toggle(s3f1, true);
	document.querySelector("#s1f2 .fruit-image").classList.toggle(s1f2, true);
	document.querySelector("#s2f2 .fruit-image").classList.toggle(s2f2, true);
	document.querySelector("#s3f2 .fruit-image").classList.toggle(s3f2, true);
	document.querySelector("#s1f3 .fruit-image").classList.toggle(s1f3, true);
	document.querySelector("#s2f3 .fruit-image").classList.toggle(s2f3, true);
	document.querySelector("#s3f3 .fruit-image").classList.toggle(s3f3, true);

	spinbutton.addEventListener("click", () => {
		spin();
		update();
	});

	betdown.addEventListener("click", () => { betchange(0); });
	betup.addEventListener("click", () => { betchange(1); });
});

function betchange(direction) {
	if (!off) return;

	if (direction === 0) {
		if (bet > 0) {
			bet--;
		}
	} else if (direction === 1) {
		if (bet < bets.length - 1) {
			bet++;
		}
	}
	update();
}

function spin() {
	if (!off) return;
	if (credits >= bets[bet]) {
		credits = credits - bets[bet];
		let spinner1 = setInterval(function () {
			spins1();
		}, 100);
		let spinner2 = setInterval(function () {
			spins2();
		}, 100);
		let spinner3 = setInterval(function () {
			spins3();
		}, 100);
		setTimeout(function () {
			clearInterval(spinner1);
		}, 2000);
		setTimeout(function () {
			clearInterval(spinner2);
		}, 3000);
		setTimeout(function () {
			clearInterval(spinner3);
		}, 4000);
		setTimeout(function () {
			checkForWin();
		}, 4500);
		setTimeout(function () {
			off = true;
		}, 5000);
		update();
	}

	off = false;
}

function spins1() {
	removeClasses(document.querySelector("#s1f3 .fruit-image"), fruitClasses);
	document.querySelector("#s1f3 .fruit-image").classList.toggle(s1f2, true);
	document.querySelector("#s1f2 .fruit-image").classList.toggle(s1f2, false);
	s1f2 = s1f1;
	document.querySelector("#s1f2 .fruit-image").classList.toggle(s1f2, true);
	document.querySelector("#s1f1 .fruit-image").classList.toggle(s1f1, false);
	s1f1 = fruits[Math.floor(Math.random() * fruits.length)];
	document.querySelector("#s1f1 .fruit-image").classList.toggle(s1f1, true);
}

function spins2() {
	removeClasses(document.querySelector("#s2f3 .fruit-image"), fruitClasses);
	document.querySelector("#s2f3 .fruit-image").classList.toggle(s2f2, true);
	document.querySelector("#s2f2 .fruit-image").classList.toggle(s2f2, false);
	s2f2 = s2f1;
	document.querySelector("#s2f2 .fruit-image").classList.toggle(s2f2, true);
	document.querySelector("#s2f1 .fruit-image").classList.toggle(s2f1, false);
	s2f1 = randomFrom(fruits);
	document.querySelector("#s2f1 .fruit-image").classList.toggle(s2f1, true);
}

function spins3() {
	removeClasses(document.querySelector("#s3f3  .fruit-image"), fruitClasses);
	document.querySelector("#s3f3 .fruit-image").classList.toggle(s3f2, true);
	document.querySelector("#s3f2 .fruit-image").classList.toggle(s3f2, false);
	s3f2 = s3f1;
	document.querySelector("#s3f2 .fruit-image").classList.toggle(s3f2, true);
	document.querySelector("#s3f1 .fruit-image").classList.toggle(s3f1, false);
	s3f1 = randomFrom(fruits);
	document.querySelector("#s3f1 .fruit-image").classList.toggle(s3f1, true);
}

function checkForWin() {
	voitto = 0;
	loppusuora = s1f2 + s2f2 + s3f2;

	if (loppusuora === "aaa" || loppusuora === "bbb" || loppusuora === "ooo") {
		voitto = voitto + bets[bet] * a3;
	}

	if (loppusuora.replace(/[^k]/g, "").length === 2) {
		voitto = voitto + bets[bet] * k2;
	}

	if (loppusuora.includes("kkk")) {
		voitto = voitto + bets[bet] * k3;
	}

	if (loppusuora.replace(/[^t]/g, "").length === 2) {
		voitto = voitto + bets[bet] * t2;
	}

	if (loppusuora.includes("ttt")) {
		voitto = voitto + bets[bet] * t3;
	}

	if (loppusuora.replace(/[^p]/g, "").length === 1) {
		voitto = voitto + bets[bet] * p1;
	}

	if (loppusuora.replace(/[^p]/g, "").length === 2) {
		voitto = voitto + bets[bet] * p2;
	}

	if (loppusuora.replace(/[^p]/g, "").length === 3) {
		voitto = voitto + bets[bet] * p3;
	}

	if (voitto > 0) {
		credits = credits + voitto;
		winspan.textContent = "Voitit " + voitto + " kolikkoa!";
	} else {
		winspan.textContent = "Ei voittoa.";
	}

	update();
}

function update() {
	spinvalue.textContent = credits.toFixed(1);
	betvalue.textContent = bets[bet];
	document.getElementById("p3p").textContent = p3 * bets[bet];
	document.getElementById("p2p").textContent = p2 * bets[bet];
	document.getElementById("p1p").textContent = p1 * bets[bet];
	document.getElementById("p3t").textContent = t3 * bets[bet];
	document.getElementById("p2t").textContent = t2 * bets[bet];
	document.getElementById("p3k").textContent = k3 * bets[bet];
	document.getElementById("p2k").textContent = k2 * bets[bet];
	document.getElementById("p3b").textContent = b3 * bets[bet];
	document.getElementById("p3o").textContent = o3 * bets[bet];
	document.getElementById("p3a").textContent = a3 * bets[bet];
}