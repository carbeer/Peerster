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
  msg = new Message($("#message").val(), undefined, undefined, undefined, undefined, undefined, undefined);
  console.log(JSON.stringify(msg));

  $.ajax({
    url: "/message",
    type: "POST",
    data: JSON.stringify(msg),
    contentType: "application/json",
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
  msg = new Message(undefined, undefined, undefined, undefined, undefined, undefined, $("#peer").val());

  $.ajax({
    url: "/node",
    type: "POST",
    data: JSON.stringify(msg),
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
  msg = new Message(undefined, undefined, $('input[type=file]').val().split('\\').pop(), undefined, undefined, undefined);
  $.ajax({
    async: false,
    url: '/file',
    type: 'POST',
    data: JSON.stringify(msg),
    success: function (data) {
      $('input[type=file]').val('');
    }
  });
}

function sendPrivateMessage() {
  msg = new Message($("#privateMessage").val(), this.privateMessagePeer, undefined, undefined, undefined, undefined);

  $.ajax({
    async: false,
    url: "/privateMessage",
    type: "POST",
    data: JSON.stringify(msg),
    success: function (data) {
      $("#privateMessage").val("Type your message here...");
    },
  });
  getPrivateMessages();
}

function sendDownloadRequest() {
  msg = new Message(undefined, $('#downloadOrigin option:selected').val(), $('#fileName').val(), $("#downloadHash").val(), undefined, undefined);

  $.ajax({
    async: false,
    url: "/download",
    type: "POST",
    data: JSON.stringify(msg),
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

Array.prototype.diff = function(a) {
  return this.filter(function(i) {return a.indexOf(i) < 0;});
};

class Message {
  constructor(text, destination, filename, request, keywords, budget, peer) {
      this.text = text;
      this.destination = destination;
      this.filename = filename;
      this.request = request;
      this.keywords = keywords;
      this.budget = budget;
      this.peer = peer;
  }
}