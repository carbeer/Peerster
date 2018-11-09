$(document).ready(function () {
  getMessages();
  getPeers();
  getKnownOrigins();
  fetchId();

  $("#sendMessage").on("click", function () {
    sendMessage();
  });

  $("#addNewPeer").on("click", function () {
    addPeer();
  });

  $('#uploadButton').on('click', function(){ 
    saveFile();
  });
});

function fetchId() {
  $.ajax({
    url: "/id",
    type: "GET",
    success: function (data) {
      id = JSON.parse(data)
      $("#id").val(id);
    },
  });
}

function sendMessage() {
  $.ajax({
    url: "/message",
    type: "POST",
    data: { message: $("#message").val() },
    success: function (data) {
      $("#message").val("Type your message here...");
    },
  });
  getMessages();
}

function getMessages() {
  $.ajax({
    url: '/message',
    type: "GET",
    success: function (data) {
      messages = JSON.parse(data);
      $('#chat').val(messages);
    },
    complete: function () {
      // Schedule the next request when the current one's complete
      setTimeout(getMessages, 5000);
    }
  });
}

function addPeer() {
  $.ajax({
    url: "/node",
    type: "POST",
    data: { peer: $("#peer").val() },
    success: function (data) {
      $("#peer").val("Type the peer here...");
    },
  });
  getPeers();
}

function getPeers() {
  $.ajax({
    url: "/node",
    type: "GET",
    success: function (data) {
      peers = JSON.parse(data);
      $("#peers").val(peers);
    },
  });
}

function getKnownOrigins() {
  $.ajax({
    url: "/origins",
    type: "GET",
    success: function (data) {
      origins = JSON.parse(origins)
      $("#origins").val(id);
    },
  });
}

function saveFile() {
  var fileName = $('#fileForm').val().split('\\').pop();
  $.ajax({
      url: '/file',  
      type: 'POST',
      data: { filename: fileName }
  });
}