$(document).ready(function () {
  const SERVER_ADDRESS = "http://localhost:8888/dig"
  $.get(SERVER_ADDRESS,{
    "time":gettime(),
    "ip":getip(),
    "url":geturl(),
    "refer":getrefer(),
    "agent":getuser_agent(),
  })
});

function gettime() {
  var nowDate = new Date();
  return nowDate.toLocaleString();
}
function geturl() {
  return window.location.href;
}
function getip() {
  return returnCitySN["cip"] + "," + returnCitySN["cname"];
}
function getrefer() {
  return document.referrer;
}
function getcookie() {
  return document.cookie;
}
function getuser_agent() {
  return navigator.userAgent;
}

function loadXMLDoc() {
  var xmlhttp;
  if (window.XMLHttpRequest) {
    xmlhttp = new XMLHttpRequest();
  } else {
    xmlhttp = new ActiveXObject("Microsoft.XMLHTTP");
  }
  xmlhttp.onreadystatechange = function () {
    if (xmlhttp.readyState == 4 && xmlhttp.status == 200) {
      //alert(xmlhttp.responseText);
    }
  };
  xmlhttp.open("POST", "http://analysis.wml.com:8088/log.php", true);
  xmlhttp.setRequestHeader("Content-type", "application/x-www-form-urlencoded");
  xmlhttp.send(
    `time=${gettime()}&ip=${getip()}&url=${geturl()}&refer=${getrefer()}&user_agent=${getuser_agent()}&cookie=${getcookie()}`
  );
}
