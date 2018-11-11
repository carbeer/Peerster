$(document).ready(function () {
  getMessages();
  getPeers();
  getKnownOrigins();
  fetchId();

  $("#sendMessage").on("click", function () {
    sendMessage();
  });

  $("#sendPrivateMessage").on("click", function () {
    sendPrivateMessage();
  });

  $("#addNewPeer").on("click", function () {
    addPeer();
  });

  $('#uploadButton').on('click', function () {
    saveFile();
  });

  $('#privateMessageOrigin').on("dblclick", "option", function () {
    openPrivateDialog($('#privateMessageOrigin option:selected').val());
  });

  $('#sendDownloadRequest').on('click', function () {
    sendDownloadRequest();
  })
});

var privateMessagePeer = "";

function fetchId() {
  $.ajax({
    url: "/id",
    type: "GET",
    success: function (data) {
      id = JSON.parse(data)
      $("#id").append(id);
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
      messages = orderKnownMessages(messages, $('#chat').val());
      $('#chat').val(messages);
    },
    complete: function () {
      setTimeout(getMessages, 5000);
    }
  });
}

function getPrivateMessages() {
  $.ajax({
    url: '/privateMessage?peer=' + privateMessagePeer,
    type: "GET",
    success: function (data) {
      messages = JSON.parse(data);
      messages = orderKnownMessages(messages, $('#privateChat').val());
      $('#privateChat').val(messages);
    },
    complete: function () {
      setTimeout(getPrivateMessages, 5000);
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
    async: false,
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
    async: false,
    url: "/origins",
    type: "GET",
    success: function (data) {
      origin = JSON.parse(data).split("\n")
      $('.selectOrigin').each(function () {
        $(this).empty()
        origin.forEach(elem => {
          if (elem != "") {
            $(this).append($('<option>', {
              id: elem,
              value: elem,
              text: elem
            }));
          }
        });
      });
    },
    complete: function () {
      setTimeout(getKnownOrigins, 5000);
    }
  });
}

function saveFile() {
  var fileName = $('input[type=file]').val().split('\\').pop();
  $.ajax({
    async: false,
    url: '/file',
    type: 'POST',
    data: { filename: fileName },
    success: function (data) {
      $('input[type=file]').val('');
    }
  });
}

function sendPrivateMessage() {
  $.ajax({
    async: false,
    url: "/privateMessage",
    type: "POST",
    data: {
      message: $("#privateMessage").val(),
      destination: this.privateMessagePeer,
    },
    success: function (data) {
      console.log($("#privateMessage").val());
      $("#privateMessage").val("Type your message here...");
    },
  });
  getPrivateMessages();
}

function sendDownloadRequest() {
  $.ajax({
    async: false,
    url: "/download",
    type: "POST",
    data: {
      destination: $('#downloadOrigin option:selected').val(),
      request: $("#downloadHash").val(),
      filename: $('#fileName').val(),
    },
    success: function (data) {
      $('#fileName').val('Desired file name...');
      $("#downloadHash").val('Insert the meta hash...');
    }
  });
}

function openPrivateDialog(origin) {
  this.privateMessagePeer = origin;
  $('#privateMessagePeer').empty().append(origin)
  document.getElementById("privateMessageDialog").style.display = "block";
  getPrivateMessages();
}

Array.prototype.diff = function(a) {
  return this.filter(function(i) {return a.indexOf(i) < 0;});
};

function orderKnownMessages(newMessages, orderedMessages) {
  if (orderedMessages == null) {
    return newMessages;
  }
  var splitOrdered = orderedMessages.split("\n");
  var splitNew = newMessages.split("\n");

  var real = splitNew.diff(splitOrdered);
  
  real.forEach(elem => {
    orderedMessages = orderedMessages + elem  + "\n";
  })
  return orderedMessages;
}

