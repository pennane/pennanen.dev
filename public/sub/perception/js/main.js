var spacepress, x, y,
   savedxyla = 350,
   savedxala = 350,
   canvas = document.getElementById("canvas"),
   ctx = canvas.getContext('2d'),
   rect = canvas.getBoundingClientRect(),
   reftumma = "#464646",
   refvaalee = "#aaaaaa",
   tummacolor = reftumma,
   vaaleecolor = refvaalee;

$('input[type=radio]').click(function() {
   console.log(this.value);
   console.log(typeof(this.value));
   if (this.value === "1") {
      tummacolor = reftumma,
         vaaleecolor = refvaalee;
   } else if (this.value === "2") {
      tummacolor = refvaalee,
         vaaleecolor = refvaalee;
   } else if (this.value === "3") {
      tummacolor = reftumma,
         vaaleecolor = reftumma;
   }
   resetCanvas();
});
$('#canvas').click(function() {
   getPosition(event);
});
$('#canvas').mousedown(function() {
   spacepress = 0;
   $("#status").text("Drawing..");
   $(document).mousemove(function(e) {
      if (spacepress === 0) {
         getPosition(event);
      }
   });
});
$(document).mouseup(function() {
   spacepress = 1;
   $("#status").text("Idle");
   return false;
});

function getPosition(event) {
   x = event.clientX - rect.left;
   y = event.clientY - rect.top;
   rectmove(x, y, event);
}

function rectmove(x, y, event) {
   if (($('#canvas').css("width") === "400px")) {
      x = x * 2;
      y = y * 2;
   }
   if (y < 300) {
      ctx.clearRect(0, 0, 1000, 300);
      if (event.clientX > rect.right) {
         ctx.fillStyle = vaaleecolor;
         ctx.fillRect(750, 125, 100, 100);
         savedxyla = 750;
      } else if (event.clientX < rect.left) {
         ctx.fillStyle = vaaleecolor;
         ctx.fillRect(-50, 125, 100, 100);
         savedxyla = 0;
      } else {
         ctx.fillStyle = vaaleecolor;
         ctx.fillRect(x - 50, 125, 100, 100);
         savedxyla = x - 50;
      };
   } else {
      ctx.clearRect(0, 300, 1000, 300);
      if (event.clientX > rect.right) {
         ctx.fillStyle = tummacolor;
         ctx.fillRect(750, 375, 100, 100);
         savedxala = 750;
      } else if (event.clientX < rect.left) {
         ctx.fillStyle = tummacolor;
         ctx.fillRect(-50, 375, 100, 100);
         savedxala = 0;
      } else {
         ctx.fillStyle = tummacolor;
         ctx.fillRect(x - 50, 375, 100, 100);
         savedxala = x - 50;
      };
   }
}

function resetCanvas() {
   ctx.clearRect(0, 0, 1000, 1000);
   ctx.fillStyle = vaaleecolor;
   ctx.fillRect(savedxyla, 125, 100, 100);
   ctx.fillStyle = tummacolor;
   ctx.fillRect(savedxala, 375, 100, 100);
}
$(document).ready(function() {
   ctx.fillStyle = vaaleecolor;
   ctx.fillRect(350, 125, 100, 100);
   ctx.fillStyle = tummacolor;
   ctx.fillRect(350, 375, 100, 100);
});