let modelUrl = "assets/thonk-model.json";
if (window.location.hash) {
    if (window.location.hash === "#real") {
        modelUrl = "assets/thonk-model-real.json";
    }
}

let model;
let loaded = false;
let container = document.getElementById("scene");
let containerWidth = 600;
let containerHeight = 600;
let scene = new THREE.Scene();
let camera = new THREE.PerspectiveCamera(800, containerWidth / containerHeight, 0.1, 1000);
let renderer = new THREE.WebGLRenderer({
    alpha: true
});
let rotx = 0,
    roty = 0,
    rotz = 0;
let spinner = document.getElementById("loaderparent")
let loadedpercent = document.getElementById("loaded")

document.body.appendChild(container);
renderer.setSize(containerWidth, containerHeight);
container.appendChild(renderer.domElement);

let loader = new THREE.ObjectLoader();

loader.load(
    modelUrl,

    (obj) => {
        console.log("Model loaded");
        model = obj;
        scene.add(obj);
        obj.applyMatrix(new THREE.Matrix4().makeTranslation(0, 0, 0));
        loaded = true;
		spinner.style.display = "none"

    },
    (xhr) => {
		if (!xhr.total) return;
		let percent = parseInt(xhr.loaded / xhr.total * 100)
		console.log(percent+"% loaded")
        loadedpercent.textContent = percent + "%"
    },

    (err) => {
        console.error('An error happened');
    }
);



scene.add(camera);
let spotLight = new THREE.PointLight(0xffffff);
let spotLight2 = new THREE.PointLight(0xffffff);
spotLight.position.set(-6, 6, 4);
camera.add(spotLight);


camera.position.z = 8;
camera.position.y = 0;

function animate() {
    requestAnimationFrame(animate);
    if (loaded) {
        model.rotation.x += rotx;
        model.rotation.y += roty;
        model.rotation.z += rotz;
        renderer.render(scene, camera);
    }

}
animate();

let xrot = document.getElementById("xrot")
let yrot = document.getElementById("yrot")
let zrot = document.getElementById("zrot")

let resetx = document.getElementById("resetx")
let resety = document.getElementById("resety")
let resetz = document.getElementById("resetz")

let xval = document.getElementById("xval")
let yval = document.getElementById("yval")
let zval = document.getElementById("zval")

xrot.addEventListener('input', (event) => {
	rotx = parseFloat(event.target.value);
	xval.textContent = (String(event.target.value))
})

yrot.addEventListener('input', (event) => {
	roty = parseFloat(event.target.value);
	yval.textContent = (String(event.target.value))
})

zrot.addEventListener('input', (event) => {
	rotz = parseFloat(event.target.value);
	zval.textContent = (String(event.target.value))
})

resetx.addEventListener('click', () => {
	model.rotation.x = 0;
    rotx = 0;
   	xval.textContent = "0"
    xrot.value = 0;
} )

resety.addEventListener('click', () => {
	model.rotation.y = 0;
    roty = 0;
   	yval.textContent = "0"
    yrot.value = 0;
} )

resetz.addEventListener('click', () => {
	model.rotation.z = 0;
    rotz = 0;
   	zval.textContent = "0"
    zrot.value = 0;
} )