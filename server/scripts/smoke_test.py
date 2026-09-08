#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
简聊 SimpleChat 全链路冒烟测试(零第三方依赖)
覆盖: 注册->登录->加好友->单聊->建群->群聊->传文件->搜索 -> WebSocket实时收发
用法: python smoke_test.py
前提: MySQL/Redis/MinIO 容器已启动, Go服务端已监听 8080
"""
import base64, hashlib, json, os, socket, struct, sys, time, urllib.request, urllib.error

BASE = "http://127.0.0.1:8080/api"
passed, failed = [], []

def req(method, path, body=None, token=None, timeout=15):
    url = BASE + path
    data = None
    headers = {}
    if body is not None:
        data = json.dumps(body).encode("utf-8")
        headers["Content-Type"] = "application/json"
    if token:
        headers["Authorization"] = "Bearer " + token
    import urllib.request as ur
    r = ur.Request(url, data=data, headers=headers, method=method)
    try:
        with ur.urlopen(r, timeout=timeout) as resp:
            return json.loads(resp.read().decode("utf-8"))
    except urllib.error.HTTPError as e:
        try:
            return json.loads(e.read().decode("utf-8"))
        except Exception:
            return {"code": e.code, "msg": str(e)}

def upload_file(token, field, name, content):
    boundary = "----smoke" + os.urandom(8).hex()
    body = b""
    body += ('--%s\r\nContent-Disposition: form-data; name="%s"; filename="%s"\r\n'
             'Content-Type: application/octet-stream\r\n\r\n' % (boundary, field, name)).encode()
    body += content
    body += ('\r\n--%s--\r\n' % boundary).encode()
    r = urllib.request.Request(BASE + "/files/upload?type=" + ("image" if field == "image" else "file"),
                               data=body, method="POST",
                               headers={"Content-Type": "multipart/form-data; boundary=" + boundary,
                                        "Authorization": "Bearer " + token})
    with urllib.request.urlopen(r, timeout=30) as resp:
        return json.loads(resp.read().decode("utf-8"))

# ---------- 极简 WebSocket 客户端 (RFC6455) ----------
class WS:
    def __init__(self, token, timeout=10):
        host = "127.0.0.1"; port = 8080
        self.sock = socket.create_connection((host, port), timeout=timeout)
        self.sock.settimeout(timeout)
        key = base64.b64encode(os.urandom(16)).decode()
        http = ("GET /ws?token=%s HTTP/1.1\r\nHost: %s:%d\r\n"
                "Upgrade: websocket\r\nConnection: Upgrade\r\n"
                "Sec-WebSocket-Key: %s\r\nSec-WebSocket-Version: 13\r\n\r\n"
                % (urllib.parse.quote(token), host, port, key))
        self.sock.sendall(http.encode())
        # 读响应头直到空行
        buf = b""
        while b"\r\n\r\n" not in buf:
            buf += self.sock.recv(4096)
        head = buf.split(b"\r\n\r\n", 1)[0].decode("latin-1")
        if " 101 " not in head:
            raise RuntimeError("WS握手失败: " + head.split("\r\n")[0])
        self.buffer = buf.split(b"\r\n\r\n", 1)[1]

    def send_text(self, obj):
        payload = json.dumps(obj).encode("utf-8")
        mask = os.urandom(4)
        n = len(payload)
        frame = bytearray([0x81])
        if n < 126:
            frame.append(0x80 | n)
        elif n < 65536:
            frame.append(0x80 | 126); frame += struct.pack(">H", n)
        else:
            frame.append(0x80 | 127); frame += struct.pack(">Q", n)
        frame += mask
        frame += bytes(b ^ mask[i % 4] for i, b in enumerate(payload))
        self.sock.sendall(frame)

    def recv(self):
        """读取并返回一条完整text帧, 无数据时抛socket.timeout"""
        if self.buffer:
            data = self.buffer; self.buffer = b""
        else:
            data = self.sock.recv(65535)
        if not data:
            return None
        # 循环直到一个完整帧(简单起见假设每帧独立TCP段)
        b0, b1 = data[0], data[1]
        n = b1 & 0x7F
        idx = 2
        if n == 126:
            n = struct.unpack(">H", data[idx:idx+2])[0]; idx += 2
        elif n == 127:
            n = struct.unpack(">Q", data[idx:idx+8])[0]; idx += 8
        masked = b1 & 0x80
        if masked:
            mkey = data[idx:idx+4]; idx += 4
            payload = bytes(b ^ mkey[i % 4] for i, b in enumerate(data[idx:idx+n]))
        else:
            payload = data[idx:idx+n]
        return json.loads(payload.decode("utf-8"))

    def close(self):
        try: self.sock.close()
        except Exception: pass

import urllib.parse as urlparse

def check(name, cond, detail=""):
    if cond:
        passed.append(name); print("  [PASS] %s %s" % (name, detail))
    else:
        failed.append(name); print("  [FAIL] %s %s" % (name, detail))

print("== 准备三个测试用户 ==")
users = {}
for acct, nick in [("alice", "Alice"), ("bob", "Bob"), ("carol", "Carol")]:
    lr = req("POST", "/auth/login", {"account": acct, "password": "pass1234"})
    if lr.get("data") and lr["data"].get("token"):
        users[acct] = lr["data"]; print("  登录 %s -> uid=%s" % (acct, lr["data"]["user"]["user_id"]))
        continue
    rg = req("POST", "/auth/register", {"account": acct, "nickname": nick, "password": "pass1234"})
    if rg.get("data") and rg["data"].get("token"):
        users[acct] = rg["data"]; print("  注册 %s -> uid=%s" % (acct, rg["data"]["user"]["user_id"]))
    else:
        print("  !!! %s 登录/注册异常 login=%s register=%s" % (acct, lr, rg))
A, B, C = users["alice"], users["bob"], users["carol"]
check("注册并登录三用户", all(u.get("user") for u in users.values()))

print("== 加好友 (Bob 申请加 Alice, Alice 同意) ==")
r = req("POST", "/contacts/apply", {"to_id": B["user"]["user_id"]}, token=A["token"])
r = req("POST", "/contacts/accept", {"from_id": A["user"]["user_id"]}, token=B["token"])
friends = req("GET", "/contacts", token=B["token"]).get("data", [])
check("好友申请/同意", any(f["user_id"] == A["user"]["user_id"] for f in friends), "Bob好友数=%d" % len(friends))

print("== 建立单聊会话 (Alice <-> Bob) ==")
r = req("POST", "/conversations/single", {"peer_id": B["user"]["user_id"]}, token=A["token"])
conv_single = r["data"]["conversation_id"]
check("单聊会话建立", conv_single > 0, "conv=%s" % conv_single)

print("== WebSocket: Alice 发单聊文本 -> Bob 实时收到 ==")
wa = WS(A["token"]); wb = WS(B["token"])
wa.send_text({"action": "send", "conversation_id": conv_single, "type": 0, "content": "hello bob, 这是一条实时消息"})
got = None
deadline = time.time() + 8
while time.time() < deadline:
    try: m = wb.recv()
    except socket.timeout: break
    if m and m.get("action") == "message" and m["data"]["conversation_id"] == conv_single:
        got = m["data"]; break
check("WS 单聊实时送达", got is not None and "hello bob" in (got.get("content") or ""), "content=%s" % (got.get("content") if got else None))

print("== 已读回执 ==")
r = req("POST", "/conversations/%s/read" % conv_single, {}, token=B["token"])
convs = req("GET", "/conversations", token=B["token"]).get("data", [])
c1 = next((c for c in convs if c["conversation_id"] == conv_single), None)
check("会话列表含单聊", c1 is not None)
check("未读数已清零", c1 is not None and c1["unread_count"] == 0, "unread=%s" % (c1["unread_count"] if c1 else "?"))

print("== 创建群聊 (Alice 建群含 Bob, Carol) ==")
r = req("POST", "/groups", {"name": "项目小组", "announcement": "欢迎", "member_ids": [B["user"]["user_id"], C["user"]["user_id"]]}, token=A["token"])
g = r["data"]
check("建群", g and g["group_id"] > 0, "group=%s" % (g["group_id"] if g else g))
groups = req("GET", "/groups/mine", token=B["token"]).get("data", [])
check("Bob 群列表含群", any(x["group_id"] == g["group_id"] for x in groups))
# 从会话列表定位群会话ID
convs = req("GET", "/conversations", token=A["token"]).get("data", [])
conv_group = next((c["conversation_id"] for c in convs if c.get("group_id") == g["group_id"]), None)
check("定位群会话", conv_group is not None, "conv=%s" % conv_group)

print("== WS: 群聊消息 -> Bob 与 Carol 实时收到 ==")
wc = WS(C["token"])
wa.send_text({"action": "send", "conversation_id": conv_group, "type": 0, "content": "开会了, 大家注意"})
gotb = gotc = None
deadline = time.time() + 8
while time.time() < deadline:
    for (w, nm) in [(wb, "bob"), (wc, "carol")]:
        try: m = w.recv()
        except socket.timeout: continue
        if m and m.get("action") == "message" and m["data"]["conversation_id"] == conv_group:
            if nm == "bob" and gotb is None and "开会" in (m["data"].get("content") or ""): gotb = m["data"]
            if nm == "carol" and gotc is None and "开会" in (m["data"].get("content") or ""): gotc = m["data"]
check("群聊送达 Bob", gotb is not None)
check("群聊送达 Carol", gotc is not None)

print("== 文件传输 ==")
r = upload_file(A["token"], "file", "demo.txt", "SimpleChat 文件内容 !!! smoke".encode("utf-8"))
fu = r["data"]
check("上传文件返回URL", fu.get("file_url") and fu.get("file_id"), fu.get("file_url"))
wa.send_text({"action": "send", "conversation_id": conv_single, "type": 2, "content": "demo.txt", "media_url": fu["file_url"]})
gotf = None; deadline = time.time() + 8
while time.time() < deadline:
    try: m = wb.recv()
    except socket.timeout: break
    if m and m.get("action") == "message" and m["data"].get("type") == 2 and m["data"]["conversation_id"] == conv_single:
        gotf = m["data"]; break
check("文件消息实时送达且带URL", gotf is not None and gotf.get("media_url"), "url=%s" % (gotf.get("media_url") if gotf else None))

print("== 消息搜索 ==")
r = req("GET", "/search/messages?keyword=%s" % urlparse.quote("hello bob"), token=B["token"])
msgs = r.get("data", [])
check("按关键词搜到消息", any("hello bob" in (m.get("content") or "") for m in msgs), "命中=%d" % len(msgs))

print("== 会话标记置顶/免打扰 ==")
req("PUT", "/conversations/%s/top?top=1" % conv_single, {}, token=A["token"])
req("PUT", "/conversations/%s/mute?mute=1" % conv_single, {}, token=A["token"])
convs = req("GET", "/conversations", token=A["token"]).get("data", [])
ce = next((c for c in convs if c["conversation_id"] == conv_single), {})
check("置顶生效", ce.get("is_top") == 1)
check("免打扰生效", ce.get("is_mute") == 1)

for w in (wa, wb, wc): w.close()
print("\n==== 结果: %d 通过, %d 失败 ====" % (len(passed), len(failed)))
if failed:
    print("失败项:", failed); sys.exit(1)
print("全部通过 ✔")