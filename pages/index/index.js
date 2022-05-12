test = () => {
	log('test');
};
//document.onfocus = () => {log("focus")};
//document.onblur = () => {log("blur")};

//document.onpointerleave = () => {log("leave")};
//document.onpointerenter = () => {log("enter")};

document.addEventListener('DOMContentLoaded', () => {
	document.onpointerout;
	window.addEventListener('mouseover', (event) => {
		if (event.relatedTarget === null) {
			console.log('enter');
		}
	});
	window.addEventListener('mouseout', (event) => {
		if (event.relatedTarget === null) {
			console.log('leave');
		}
	});
	window.addEventListener('focus', (event) => {
		console.log('focus');
	});
	window.addEventListener('blur', (event) => {
		console.log('blur');
	});
	window.addEventListener('resize', (event) => {
		console.log('resize');
	});
	console.log('Listening');
	log('Listening');
});

function toggleMessage(item) {
	const message = document.getElementsByClassName(item.id)[0];
	if (message.classList.contains('selected-item')) {
		document.getElementsByClassName(item.id)[0].classList.remove('selected-item');
		ToggleMessage(item.id.split('-')[1], false)
	} else {
		document.getElementsByClassName(item.id)[0].classList.add('selected-item');
		ToggleMessage(item.id.split('-')[1], true)
	}
}
