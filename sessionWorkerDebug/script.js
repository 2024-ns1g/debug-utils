function logMessage(role, message) {
  var logElem = document.getElementById('log-' + role);
  logElem.textContent += message + "\n";
  logElem.scrollTop = logElem.scrollHeight;
}

async function startConnection(role) {
  var baseURL = document.getElementById('baseURL').value.trim();
  var otpInput = document.getElementById('otp-' + role);
  var otp = otpInput.value.trim();
  if (!otp || otp.length !== 6 || !/^[0-9]{6}$/.test(otp)) {
    alert('有効な6桁のOTPを入力してください');
    return;
  }
  logMessage(role, "OTP認証開始…");
  try {
    // OTP認証用URL
    var verifyUrl = baseURL + "/session/" + role + "/verify";
    var res = await fetch(verifyUrl, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ otp: otp })
    });
    if (!res.ok) {
      logMessage(role, "OTP認証失敗: " + res.statusText);
      return;
    }
    var data = await res.json();
    var sessionId = data.sessionId;
    var token = data.token;
    var aggregatorUrl = data.aggregatorUrl;
    logMessage(role, "OTP認証成功");
    logMessage(role, "sessionId: " + sessionId);
    logMessage(role, "token: " + token);
    logMessage(role, "aggregatorUrl: " + aggregatorUrl);
    
    // WebSocket接続用URL作成
    var wsUrl = aggregatorUrl + "/" + role + "?sessionId=" + encodeURIComponent(sessionId);
    logMessage(role, "WebSocket接続: " + wsUrl);
    var ws = new WebSocket(wsUrl);
    
    ws.onopen = function() {
      logMessage(role, "WebSocket接続確立");
      // 認証用メッセージ送信
      var dataObj = { token: token };
      if (role === "agent") {
        var agentName = document.getElementById('agentName').value.trim();
        var agentType = document.getElementById('agentType').value.trim();
        dataObj.agentName = agentName;
        dataObj.agentType = agentType;
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
    };
    
    ws.onclose = function() {
      logMessage(role, "WebSocket接続終了");
    };
    
  } catch (error) {
    logMessage(role, "エラー: " + error);
  }
}
