(function() {
  var copylinkbtn = document.getElementById("copy-link-btn"),
      copylink = document.getElementById("copy-link-wrapper"),
      overlay = document.getElementById("overlay");

    var url = "http://url"
  copylinkbtn.addEventListener("click", function(e) {
    e.preventDefault();
    
    var error = document.getElementsByClassName('error');
      
    while (error[0]) {
      error[0].parentNode.removeChild(error[0]);
    }

    document.body.className += ' active';
    
    copylink.children[1].value = url;
    copylink.children[1].focus();
    copylink.children[1].select();
    return (false);
  }, false);

  overlay.addEventListener("click", function(e) {
    e.preventDefault();
    document.body.className = '';
    return (false);
  }, false);

  copylink.children[2].addEventListener("keydown", function(e) {

    var error = document.getElementsByClassName('error');

    while (error[0]) {
      error[0].parentNode.removeChild(error[0]);
    }

    setTimeout(function() {

      if((e.metaKey || e.ctrlKey) && e.keyCode === 67 && isTextSelected(copylink.children[2])) {
        document.body.className = '';
      } else if((e.metaKey || e.ctrlKey) && e.keyCode === 67 && isTextSelected(copylink.children[2]) === false) {
        var error = document.createElement('span');
        error.className = 'error';
        var errortext = document.createTextNode('The link was not copied, make sure the entire text is selected.');
        
        error.appendChild(errortext);
        copylink.appendChild(error);
      }
    }, 100);

    function isTextSelected(input) {
      if (typeof input.selectionStart == "number") {
        return input.selectionStart == 0 && input.selectionEnd == input.value.length;
      } else if (typeof document.selection != "undefined") {
        input.focus();
        return document.selection.createRange().text == input.value;
      }
    }
  }, false);
})();
