// 各パネルのWebSocketインスタンスを保持
var sockets = {};

// ローカルストレージのキー名を一元管理するためのヘルパー
function getHistoryKey(role) {
  return "messageHistory_" + role;
}

// 履歴を取得する
function getMessageHistory(role) {
  var str = localStorage.getItem(getHistoryKey(role));
  if (str) {
    try {
      return JSON.parse(str);
    } catch (e) {
      // パースエラー時は空配列を返す
      return [];
    }
  }
  return [];
}

// 履歴を保存する
function saveMessageHistory(role, history) {
  localStorage.setItem(getHistoryKey(role), JSON.stringify(history));
}

function updateHistoryPanel(role) {
  var historyContainer = document.getElementById("history-" + role);
  if (!historyContainer) return;

  var history = getMessageHistory(role);
  if (history.length === 0) {
    historyContainer.textContent = "まだメッセージ履歴はありません。";
    return;
  }

  // いったん空にしてから追加
  historyContainer.innerHTML = "";

  history.forEach(function(msg, index) {
    var escapedReqType = escapeHtml(msg.requestType);
    var escapedData = JSON.stringify(msg.data);

    // 履歴アイテム全体をラップするdiv
    var itemDiv = document.createElement("div");
    itemDiv.className = "history-item";

    // 左側のテキスト部分
    var infoDiv = document.createElement("div");
    infoDiv.className = "history-info";
    infoDiv.textContent =
      "No." + (index + 1) + " [requestType]: " + escapedReqType + " [data]: " + escapedData;

    // 右側の「再送」ボタン
    var actionDiv = document.createElement("div");
    actionDiv.className = "history-action";
    actionDiv.innerHTML =
      "<button onclick='loadHistoryMessage(\"" + role + "\"," + index + ")'>再送</button>";

    // アイテムをまとめて格納
    itemDiv.appendChild(infoDiv);
    itemDiv.appendChild(actionDiv);

    historyContainer.appendChild(itemDiv);
  });
}

// 特殊文字のエスケープ
function escapeHtml(text) {
  return text
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#039;");
}

// 履歴から選んだメッセージを入力欄に反映
function loadHistoryMessage(role, index) {
  var history = getMessageHistory(role);
  var msg = history[index];
  if (!msg) return;

  var reqTypeElem = document.getElementById("requestType-" + role);
  var dataElem = document.getElementById("data-" + role);
  reqTypeElem.value = msg.requestType;
  dataElem.value = JSON.stringify(msg.data, null, 2);
}

// 履歴パネルの表示・非表示を切り替え
function toggleHistoryPanel(role) {
  var panel = document.getElementById("history-" + role);
  if (!panel) return;

  // 表示を更新
  if (panel.style.display === "none" || !panel.style.display) {
    panel.style.display = "block";
    updateHistoryPanel(role);
  } else {
    panel.style.display = "none";
  }
}

function logMessage(role, message) {
  var logElem = document.getElementById("log-" + role);
  logElem.textContent += message + "\n";
  logElem.scrollTop = logElem.scrollHeight;
}

async function startConnection(role) {
  var baseURL = document.getElementById("baseURL").value.trim();
  var otpInput = document.getElementById("otp-" + role);
  var otp = otpInput.value.trim();
  if (!otp || otp.length !== 6 || !/^[0-9]{6}$/.test(otp)) {
    alert("有効な6桁のOTPを入力してください");
    return;
  }
  logMessage(role, "OTP認証開始…");
  var statusElem = document.getElementById("status-" + role);
  statusElem.textContent = "接続中";
  statusElem.style.color = "green";
  try {
    var verifyUrl = baseURL + "/session/" + role + "/verify";
    var res = await fetch(verifyUrl, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ otp: otp })
    });
    if (!res.ok) {
      logMessage(role, "OTP認証失敗: " + res.statusText);
      statusElem.textContent = "未接続";
      statusElem.style.color = "red";
      return;
    }
    var data = await res.json();
    var sessionId = data.sessionId;
    var token = data.token;
    var aggregatorUrl = data.aggregatorUrl;
    logMessage(role, "OTP認証成功");

    var wsUrl = aggregatorUrl + "/" + role + "?sessionId=" + encodeURIComponent(sessionId);
    logMessage(role, "WebSocket接続: " + wsUrl);
    var ws = new WebSocket(wsUrl);

    ws.onopen = function() {
      logMessage(role, "WebSocket接続確立");
      sockets[role] = ws;
      var dataObj = { token: token };
      if (role === "agent") {
        var agentName = document.getElementById("agentName").value.trim();
        dataObj.agentName = agentName;
        dataObj.agentType = "SHOW_PRESENTATION_COMPUTER";
      }
      var msg = {
        requestType: "REGIST_" + role.toUpperCase(),
        data: dataObj
      };
      ws.send(JSON.stringify(msg));
      logMessage(role, "認証メッセージ送信: " + JSON.stringify(msg));
    };

    ws.onmessage = function(event) {
      logMessage(role, "受信: " + event.data);
    };

    ws.onerror = function(error) {
      logMessage(role, "エラー: " + error);
      statusElem.textContent = "エラー";
      statusElem.style.color = "red";
    };

    ws.onclose = function() {
      logMessage(role, "WebSocket接続終了");
      statusElem.textContent = "切断";
      statusElem.style.color = "red";
    };

  } catch (error) {
    logMessage(role, "エラー: " + error);
    statusElem.textContent = "エラー";
    statusElem.style.color = "red";
  }
}

function sendMessage(role) {
  var reqTypeElem = document.getElementById("requestType-" + role);
  var dataElem = document.getElementById("data-" + role);
  var requestType = reqTypeElem.value.trim();
  var dataText = dataElem.value.trim();
  if (!requestType) {
    alert("Request Type を入力してください");
    return;
  }
  if (!sockets[role] || sockets[role].readyState !== WebSocket.OPEN) {
    alert("WebSocketが接続されていません (" + role + ")");
    return;
  }
  var parsedData;
  try {
    parsedData = JSON.parse(dataText);
  } catch (e) {
    alert("Dataは有効なJSONオブジェクトを入力してください");
    return;
  }
  var msgObj = {
    requestType: requestType,
    data: parsedData
  };
  sockets[role].send(JSON.stringify(msgObj));
  logMessage(role, "送信: " + JSON.stringify(msgObj));

  // --- 変更点: ここではクリアしないようにするためコメントアウト ---
  // reqTypeElem.value = "";
  // dataElem.value = "";

  // 履歴保存
  var history = getMessageHistory(role);
  history.push(msgObj);
  saveMessageHistory(role, history);
}
