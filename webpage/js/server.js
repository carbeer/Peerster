$(document).ready(function () {
  getMessages();
  getPeers();
  getKnownOrigins();
  fetchId();
  getAvailableFiles();
  getPrivateFiles();

  $("#sendMessage").on("click", function (
  ) {
    sendMessage();
  });

  $("#sendPrivateMessage").on("click", function () {
    sendPrivateMessage();
  });

  $("#sendPrivateEncryptedMessage").on("click", function () {
    sendPrivateEncryptedMessage();
  });

  $("#addNewPeer").on("click", function () {
    addPeer();
  });

  $('#uploadButton').on('click', function () {
    saveFile($(this));
  });

  $('#privateUploadButton').on('click', function () {
    savePrivateFile($(this));
  });

  $("#importJSON").on("click", function () {
    uploadExportFile($(this));
  })

  $('#privateMessageOrigin').on("dblclick", "option", function () {
    openPrivateDialog($('#privateMessageOrigin option:selected').val());
  });

  $('#availableFiles').on("dblclick", "option", function () {
    // ID corresponds to metahash, value to name of the file
    downloadFile($('#availableFiles option:selected').attr('id'), $('#availableFiles option:selected').val());
  });

  $("#availablePrivateFiles").on("dblclick", "option", function() {
    downloadPrivateFile($('#availablePrivateFiles option:selected').attr('id'), $('#availablePrivateFiles option:selected').attr('value'));
  })  

  $('#searchRequest').on('click', function () {
    searchRequest();
  })

  $('#sendDownloadRequest').on('click', function () {
    sendDownloadRequest();
  })

  $('#exportJSON').on("click", function () {
    exportPrivateFiles();
  })

  $('input[type="file"]').change(function (e) {
    var fileName = e.target.files[0].name;
    $(this).next('.custom-file-label').html(fileName);
  });
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
  msg = new Message($("#message").val(), null, null, null, null, null, null, null, null, null);

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
  msg = new Message(null, null, null, null, null, null, $("#peer").val(), null, null);

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

function saveFile(el) {
  msg = new Message(null, null, $('#fileForm').val().split('\\').pop(), null, null, null, null, null, null);
  $.ajax({
    async: true,
    url: '/file',
    type: 'POST',
    data: JSON.stringify(msg),
    success: function () {
      $('#fileForm').val("");
      el.closest(':has(.custom-file-label)').find(".custom-file-label").html("Choose file...");
    }
  });
}

function savePrivateFile(el) {
  console.log($("#replications option:selected"));
  reps = parseInt($("#replications option:selected").val());
  msg = new Message(null, null, $('#privateFileForm').val().split('\\').pop(), null, null, null, null, null, reps);
  $.ajax({
    async: true,
    url: '/privateFile',
    type: 'POST',
    data: JSON.stringify(msg),
    success: function () {
      $('#privateFileForm').val("");
      el.closest(':has(.custom-file-label)').find(".custom-file-label").html("Choose file...");
      getPrivateFiles();
    }
  });
}

function getPrivateFiles() {
  $.ajax({
    async: false,
    url: '/privateFile',
    type: 'GET',
    complete: function (data) {
      map = JSON.parse(data.responseText)
      $('#availablePrivateFiles').empty()
      Object.keys(map).forEach(function (key) {
        $('#availablePrivateFiles').append($('<option>', {
          id: key,
          value: map[key],
          text: map[key]
        }));
      });
      setTimeout(getPrivateFiles, 5000);
    }
  });
}

function sendPrivateMessage() {
  msg = new Message($("#privateMessage").val(), this.privateMessagePeer, null, null, null, null, null, null, null);

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
  msg = new Message($("#privateEncryptedMessage").val(), this.privateMessagePeer, null, null, null, null, null, true, null);
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
  msg = new Message(null, $('#downloadOrigin option:selected').val(), $('#fileName').val(), $("#downloadHash, null").val(), null, null, null, null);

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
  msg = new Message(null, null, null, null, $('#searchRequestKeywords').val().split("\n, null"), -1, null, null, null);

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
  msg = new Message(null, null, name, metahash, null, null, null, null, null);
  $.ajax({
    async: false,
    url: '/download',
    type: "POST",
    data: JSON.stringify(msg),
    complete: function () {
    }
  });
}

function downloadPrivateFile(key, value) {
  msg = new Message(null, null, value, key, null, null, null, true, null);
  $.ajax ({
    async: false,
    url: '/download',
    type: "POST",
    data: JSON.stringify(msg),
    complete: function () {
    }
  });
}

function exportPrivateFiles() {
  $.ajax({
    async: false,
    url: '/export',
    type: "GET",
    complete: function (data) {
      var element = document.createElement('a');
      element.setAttribute('href', 'data:text/json;charset=utf-8,' + encodeURIComponent(data.responseText));
      element.setAttribute('download', "Private_File_Export.json");
      element.style.display = 'none';
      document.body.appendChild(element);
      element.click();
      document.body.removeChild(element);

    }
  });
}

function uploadExportFile(el) {
  msg = new Message(null, null, $('#importJSONform').val().split('\\').pop(), null, null, null, null, true, null);
  $.ajax({
    async: true,
    url: '/export',
    type: 'POST',
    data: JSON.stringify(msg),
    success: function () {
      $('#importJSONform').val("");
      el.closest(':has(.custom-file-label)').find(".custom-file-label").html("Choose file...");
      getPrivateFiles();
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
  constructor(text, destination, filename, request, keywords, budget, peer, encrypted, replications) {
    this.text = text;
    this.destination = destination;
    this.filename = filename;
    this.request = request;
    this.keywords = keywords;
    this.budget = budget;
    this.peer = peer;
    this.encrypted = encrypted;
    this.replications = replications;
    Object.keys(this).forEach((key) => (this[key] == null) && delete this[key]);
  }
}
