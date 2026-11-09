/* --- Modified WebGL checker from panic.com/transmit  ---- */
var canvas = document.createElement("canvas"),
   webglContextParams = ['webgl', 'experimental-webgl', 'webkit-3d', 'moz-webgl'],
   webglContext = null,
   i = 0;
for (i; i < webglContextParams.length; i++) {
   try {
      webglContext = canvas.getContext(webglContextParams[i]);
      if (webglContext) {
         break;
      }
   } catch (error) {
      console.log(e);
   }
}
if (!(webglContext === null)) {
   load3d();
}
/* --------------------------------------------------- */
function load3d() {
   $("#scene").css("width", "auto");
   $("#scene").css("height", "auto");
   $("#scene").css("background-image", "none");
   var modelUrl = "assets/thonk-model.json";
   var model;
   var loaded = false;
   var scalefactor = 1.19;
   var container = document.getElementById("scene");
   var containerWidth = 128;
   var containerHeight = 128;
   var scene = new THREE.Scene();
   var camera = new THREE.PerspectiveCamera(800, containerWidth / containerHeight, 0.1, 1000);
   var renderer = new THREE.WebGLRenderer({
      alpha: true
   });
   var rotx = 0,
      roty = 0.016,
      rotz = 0;
   var initbw = parseInt($("body").css("width"));
   $("#scene").append(container);
   if (initbw < 400) {
      renderer.setSize(128, 128);
   } else if (400 < initbw && initbw < 1200) {
      renderer.setSize(192, 192);
   } else if (1200 < initbw) {
      renderer.setSize(256, 256);
   }
   container.appendChild(renderer.domElement);
   var loader = new THREE.ObjectLoader();
   loader.load(
      // resource URL
      modelUrl,
      function(obj) {
         console.log("Model loaded");
         model = obj;
         scene.add(obj);
         obj.applyMatrix(new THREE.Matrix4().makeTranslation(0, 0.8, 0));
         model.scale.x = scalefactor;
         model.scale.z = scalefactor;
         model.scale.y = scalefactor;
         loaded = true;
      },
      function(err) {
         console.error('An error happened');
      });
   scene.add(camera);
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
   $(window).resize(function() {
      var bw = $("body").css("width");
      bw = parseInt(bw);
      if (bw < 400) {
         renderer.setSize(128, 128);
      } else if (400 < bw && bw < 1200) {
         renderer.setSize(192, 192);
      } else if (1200 < bw) {
         renderer.setSize(256, 256);
      }
   });
   animate();
}

$("#curYr").text((new Date()).getFullYear());