var galleryLoad = false;
var galleryOffset = 0;
var xmlhttp;
function GetXmlHttpObject(){
	if(window.XMLHttpRequest){
		// code for IE7+, Firefox, Chrome, Opera, Safari
		return new XMLHttpRequest();
	}
	if(window.ActiveXObject){
		// code for IE6, IE5
		return new ActiveXObject("Microsoft.XMLHTTP");
	}
	return null;
}
function ReverseDisplay(d){
	if(document.getElementById(d).style.display == "none"){
		document.getElementById(d).style.display = "";
	}else{
		document.getElementById(d).style.display = "none";
	}
}
function openGallery(rdisp){
	if(rdisp == true){
		ReverseDisplay('gallery');
		ReverseDisplay('galleryrefresh');
	}
	if(galleryLoad == false){
		xmlhttp = GetXmlHttpObject();
		document.getElementById('gallery').innerHTML = "rollin...";
		if(xmlhttp==null){
			alert("Failed to send XML HTTP Request!");
			return;
		}
		xmlhttp.open("GET","/?jsgallery&offset="+encodeURIComponent(galleryOffset),true);
		xmlhttp.setRequestHeader("Content-Type", "application/x-www-form-urlencoded");
		xmlhttp.send(null);
		xmlhttp.onreadystatechange = function (){
			if(xmlhttp.readyState==4){
				galleryLoad = true;
				document.getElementById('gallery').innerHTML = xmlhttp.responseText;
				if(hashData != ""){
					hashData.replace(/#/gi, "");
					displayImage(document.getElementById(hashData.replace(/#/gi, "")).href, hashData.replace(/#IMG/gi, ""));
				}
			}
		}
	}
}
var adr = "";
function displayImage(img, id){
	if(img == ""){
		return true;
	}
	adr = img;

	return false;
}
function displayImageLoaded(){

}
function displayImageOpen(){
	location.href = adr;
}