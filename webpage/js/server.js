$(document).ready(function () {
  getMessages();
  getPeers();
  getKnownOrigins();
  fetchId();
  getAvailableFiles();

  $("#sendMessage").on("click", function () {
    sendMessage();
  });

  $("#sendPrivateMessage").on("click", function () {
    sendPrivateMessage();
  });

  $("#sendPrivateEncryptedMessage").on("click", function() {
    sendPrivateEncryptedMessage();
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

  $('#availableFiles').on("dblclick", "option", function () {
    // ID corresponds to metahash, value to name of the file
    downloadFile($('#availableFiles option:selected').attr('id'), $('#availableFiles option:selected').val());
  });

  $('#searchRequest').on('click', function () {
    searchRequest();
  })

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
  msg = new Message($("#message").val(), null, null, null, null, null, null, null, null);

  $.ajax({
    url: "/message",
    type: "POST",
    data: JSON.stringify(msg),
    contentType: "application/json",
    success: function (data) {
      $("#message").val("")
    },
  });
  getMessages();
}

function getMessages() {
  $.ajax({
    url: '/message',
    type: "GET",
    success: function (data) {
      $('#chat').val(JSON.parse(data));
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
      $('#privateChat').val(JSON.parse(data));
    },
    complete: function () {
      setTimeout(getPrivateMessages, 5000);
    }
  });
}

function addPeer() {
  msg = new Message(null, null, null, null, null, null, $("#peer").val(), null);

  $.ajax({
    url: "/node",
    type: "POST",
    data: JSON.stringify(msg),
    success: function (data) {
      $("#peer").val("");
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
      updateDirectPeers(data)
    },
    complete: function () {
      setTimeout(getKnownOrigins, 5000);
    }
  });
}

function updateDirectPeers(data) {
  origin = JSON.parse(data).split("\n")
  $('#downloadOrigin').empty();
  $('#privateMessageOrigin').empty();
  origin.forEach(elem => {
    if (elem != "") {
      $('#downloadOrigin').append($('<option>', {
        id: elem,
        value: elem,
        text: elem
      }));
      $('#privateMessageOrigin').append($('<option>', {
        id: elem,
        value: elem,
        text: elem
      }));
    }
  });
}

function saveFile() {
  msg = new Message(null, null, $('input[type=file]').val().split('\\').pop(), null, null, null, null, null);
  $.ajax({
    async: true,
    url: '/file',
    type: 'POST',
    data: JSON.stringify(msg),
    success: function () {
      $('input[type=file]').val("");
    }
  });
}

function sendPrivateMessage() {
  msg = new Message($("#privateMessage").val(), this.privateMessagePeer, null, null, null, null, null, null);

  $.ajax({
    async: false,
    url: "/privateMessage",
    type: "POST",
    data: JSON.stringify(msg),
    success: function () {
      $("#privateMessage").val("");
    },
  });
  getPrivateMessages();
}

function sendPrivateEncryptedMessage() {
  msg = new Message($("#privateEncryptedMessage").val(), this.privateMessagePeer, null, null, null, null, null, true);
  $.ajax({
    async: false,
    url: "/privateMessage",
    type: "POST",
    data: JSON.stringify(msg),
    success: function () {
      $("#privateEncryptedMessage").val("");
    },
  });
  getPrivateMessages();
}


function sendDownloadRequest() {
  msg = new Message(null, $('#downloadOrigin option:selected').val(), $('#fileName').val(), $("#downloadHash, null").val(), null, null, null);

  $.ajax({
    async: false,
    url: "/download",
    type: "POST",
    data: JSON.stringify(msg),
    success: function () {
      $('#fileName').val('Desired file name...');
      $("#downloadHash").val('Insert the meta hash...');
    }
  });
}

function openPrivateDialog(origin) {
  this.privateMessagePeer = origin;
  document.getElementById("privateMessageDialog").style.display = "block";
  getPrivateMessages();
}

function getAvailableFiles() {
  $.ajax({
    async: false,
    url: "/searchRequest",
    type: "GET",
    success: function (data) {
      files = JSON.parse(data);
      $('#availableFiles').empty();
      files.forEach(elem => {
        if (elem != "") {
          $('#availableFiles').append($('<option>', {
            id: elem.metaHash,
            value: elem.fileName,
            text: elem.fileName
          }));
        }
      });
    },
    complete: function () {
      setTimeout(getAvailableFiles, 5000);
    }
  });
}

function searchRequest() {
  msg = new Message(null, null, null, null, $('#searchRequestKeywords').val().split("\n, null"), -1, null, null);

  $.ajax({
    async: false,
    url: "/searchRequest",
    type: "POST",
    data: JSON.stringify(msg),
    success: function () {
      $('#searchRequestKeywords').empty();
      getAvailableFiles();
    },
  });
}

function downloadFile(metahash, name) {
  msg = new Message(null, null, name, metahash, null, null, null, null);
  $.ajax, null({
    async: false,
    url: '/download',
    type: "POST",
    data: JSON.stringify(msg),
    complete: function () {
    }
  });
}

Array.prototype.diff = function (a) {
  return this.filter(function (i) { return a.indexOf(i) < 0; });
};

class File {
  constructor(fileName, metaHash) {
    this.fileName = fileName;
    this.metaHash = metaHash;
    Object.keys(this).forEach((key) => (this[key] == null) && delete this[key]);
  }
}

class Message {
  constructor(text, destination, filename, request, keywords, budget, peer, encrypted) {
    this.text = text;
    this.destination = destination;
    this.filename = filename;
    this.request = request;
    this.keywords = keywords;
    this.budget = budget;
    this.peer = peer;
    this.encrypted = encrypted;
    Object.keys(this).forEach((key) => (this[key] == null) && delete this[key]);
  }
}