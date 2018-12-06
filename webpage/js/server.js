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

  $('#availableFiles').on("dblclick", "option", function () {
    // ID corresponds to metahash, value to name of the file
    downloadFile($('#availableFiles option:selected').id(), $('#availableFiles option:selected').val());
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
  msg = new Message($("#message").val(), null, null, null, null, null, null);

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
  msg = new Message(null, null, null, null, null, null, $("#peer").val());

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
      $('#downloadOrigin', '#privateMessageOrigin').each(function () {
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
  msg = new Message(null, null, $('input[type=file]').val().split('\\').pop(), null, null, null);
  console.log("Saved " + JSON.stringify(msg));
  $.ajax({
    async: true,
    url: '/file',
    type: 'POST',
    data: JSON.stringify(msg),
    success: function () {
      $('input[type=file]').val('');
    }
  });
}

function sendPrivateMessage() {
  msg = new Message($("#privateMessage").val(), this.privateMessagePeer, null, null, null, null);

  $.ajax({
    async: false,
    url: "/privateMessage",
    type: "POST",
    data: JSON.stringify(msg),
    success: function () {
      $("#privateMessage").val("Type your message here...");
    },
  });
  getPrivateMessages();
}

function sendDownloadRequest() {
  msg = new Message(null, $('#downloadOrigin option:selected').val(), $('#fileName').val(), $("#downloadHash").val(), null, null);

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
  msg = new Message(null, null, null, null, $('#searchRequestKeywords').val().split("\n"), -1, null);

  $.ajax({
    async: false,
    url: "/searchRequest",
    type: "POST",
    data: JSON.stringify(msg),
    success: function () {
      $('#searchRequestKeywords').empty();
    },
    complete: function () {
      getAvailableFiles();
    }
  });
}

function downloadFile(metahash, name) {
  msg = new Message(null, null, name, metahash, null, null, null);
  $.ajax({
    async: false,
    url: '/download',
    type: "POST",
    data: JSON.stringify(msg),
    complete: function () {
      console.log("Downloaded file");
    }
  });
}

Array.prototype.diff = function (a) {
  return this.filter(function (i) { return a.indexOf(i) < 0; });
};

class File {
  constructor(fileName, metaHash, fileSize) {
    this.fileName = fileName;
    this.metaHash = metaHash;
    this.fileSize = fileSize;
    Object.keys(this).forEach((key) => (this[key] == null) && delete this[key]);
  }
}

class Message {
  constructor(text, destination, filename, request, keywords, budget, peer) {
    this.text = text;
    this.destination = destination;
    this.filename = filename;
    this.request = request;
    this.keywords = keywords;
    this.budget = budget;
    this.peer = peer;
    Object.keys(this).forEach((key) => (this[key] == null) && delete this[key]);
  }
}