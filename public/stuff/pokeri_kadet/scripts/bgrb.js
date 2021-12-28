//Original code from https://bl.ocks.org/AGdelaX/856e67e9ba5905c051db, modified slightly
var varit = [ "indigo", "violet", "red", "orange", "yellow", "green", " blue" ];
var n = 1;
window.setInterval( function() {
	document.body.style.backgroundColor = varit[ n ];
	document.getElementById( "footerbg" )
		.style.backgroundColor = varit[ n ];
	n++;
	if ( n === varit.length ) {
		n = 0;
	}
}, 5000 );