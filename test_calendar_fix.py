#!/usr/bin/env python3
"""端到端测试：日历通知系统和可见性修复验证"""
import requests
import json
import sys
import time
import base64
from datetime import datetime, timedelta
from cryptography.hazmat.primitives import serialization
from cryptography.hazmat.primitives.asymmetric import padding
from cryptography.hazmat.primitives import hashes

BASE = "https://localhost:8080/api"
VERIFY_SSL = False

def req(method, path, token=None, **kwargs):
    headers = {}
    if token:
        headers["Authorization"] = f"Bearer {token}"
    kwargs["headers"] = headers
    kwargs["verify"] = VERIFY_SSL
    r = getattr(requests, method)(f"{BASE}{path}", **kwargs)
    try:
        return r.json()
    except:
        return {"code": r.status_code, "raw": r.text}

def rsa_encrypt_password(public_key_pem, plain_password):
    """使用RSA-OAEP + SHA-256加密密码"""
    public_key = serialization.load_pem_public_key(public_key_pem.encode())
    encrypted = public_key.encrypt(
        plain_password.encode(),
        padding.OAEP(
            mgf=padding.MGF1(algorithm=hashes.SHA256()),
            algorithm=hashes.SHA256(),
            label=None
        )
    )
    return base64.b64encode(encrypted).decode()

def get_public_key():
    r = requests.get(f"{BASE}/public-key", verify=VERIFY_SSL)
    data = r.json()
    return data["data"]["public_key"]

def login(username, password, pub_key_pem):
    encrypted_pw = rsa_encrypt_password(pub_key_pem, password)
    r = req("post", "/login", json={"username": username, "password": encrypted_pw})
    if r.get("code") == 200:
        return r["data"]["token"], r["data"]["username"]
    print(f"  LOGIN FAILED for {username}: {r}")
    return None, None

def test_all():
    results = []
    
    def check(name, condition, detail=""):
        status = "PASS" if condition else "FAIL"
        results.append((name, status, detail))
        icon = "✅" if condition else "❌"
        print(f"  {icon} {name}" + (f" - {detail}" if detail and not condition else ""))
        return condition

    # 1. 登录两个账号
    print("\n=== 1. 登录测试 ===")
    pub_key = get_public_key()
    check("获取RSA公钥", pub_key is not None)
    
    ylw_token, ylw_name = login("ylw", "!Qw2!Qw2", pub_key)
    check("ylw登录", ylw_token is not None, f"name={ylw_name}")
    
    zbj_token, zbj_name = login("zbj", "admin123456", pub_key)
    check("zbj登录", zbj_token is not None, f"name={zbj_name}")
    
    if not ylw_token or not zbj_token:
        print("\n登录失败，无法继续测试")
        return results

    # 2. 清理旧数据 - 先查看zbj当前日程和通知
    print("\n=== 2. 检查zbj当前日程（应无旧日程） ===")
    zbj_calendars = req("get", "/calendars", token=zbj_token)
    print(f"  zbj当前日程数: {len(zbj_calendars.get('data', []))}")

    # 3. ylw创建包含zbj的日程（未来日程）
    print("\n=== 3. ylw创建包含zbj的未来日程 ===")
    future_start = (datetime.now() + timedelta(days=3)).replace(hour=10, minute=0, second=0, microsecond=0)
    future_end = future_start + timedelta(hours=1)
    
    create_resp = req("post", "/calendars", token=ylw_token, json={
        "title": "测试未来会议-ylw创建",
        "description": "这是ylw创建的测试日程",
        "start_time": future_start.isoformat(),
        "end_time": future_end.isoformat(),
        "is_all_day": False,
        "participants": [
            {"user_dn": "zbj", "display_name": "zbj"}
        ]
    })
    check("ylw创建日程成功", create_resp.get("code") == 200, f"resp={create_resp.get('code')}")
    calendar_id = create_resp.get("data", {}).get("id")
    check("日程ID有效", calendar_id is not None, f"id={calendar_id}")

    # 4. zbj查看日程列表（可见性测试）
    print("\n=== 4. zbj检查日程可见性 ===")
    time.sleep(1)  # 等待数据同步
    zbj_cals = req("get", "/calendars", token=zbj_token)
    zbj_cal_list = zbj_cals.get("data", [])
    found_calendar = any(c.get("id") == calendar_id for c in zbj_cal_list)
    check("zbj能看到ylw创建的日程", found_calendar, f"日程列表数={len(zbj_cal_list)}")

    # 5. zbj查看日程详情
    if calendar_id:
        print("\n=== 5. zbj查看日程详情 ===")
        detail_resp = req("get", f"/calendars/{calendar_id}", token=zbj_token)
        check("zbj能查看日程详情", detail_resp.get("code") == 200)
        detail_data = detail_resp.get("data", {})
        check("日程标题正确", detail_data.get("title") == "测试未来会议-ylw创建")
        participants = detail_data.get("participants", [])
        has_zbj = any(p.get("user_dn") == "zbj" for p in participants)
        check("日程包含zbj参与者", has_zbj, f"participants={participants}")

    # 6. zbj检查通知（应有未读通知）
    print("\n=== 6. zbj检查通知 ===")
    time.sleep(1)
    notif_resp = req("get", "/calendars/today-notifications", token=zbj_token)
    notifications = notif_resp.get("data", [])
    check("zbj收到通知", len(notifications) > 0, f"通知数={len(notifications)}")
    
    # 检查是否有对应这个日程的通知
    has_calendar_notif = any(n.get("calendar_id") == calendar_id for n in notifications)
    check("通知包含新创建的日程", has_calendar_notif)

    # 7. zbj检查未读数量
    print("\n=== 7. zbj检查未读数量 ===")
    unread_resp = req("get", "/calendars/unread-count", token=zbj_token)
    unread_count = unread_resp.get("data", {}).get("count", 0)
    check("zbj未读数量>0", unread_count > 0, f"unread={unread_count}")

    # 8. zbj标记通知为已读
    print("\n=== 8. zbj标记通知已读 ===")
    if notifications:
        notif_id = notifications[0]["id"]
        mark_resp = req("put", f"/calendars/notifications/{notif_id}/read", token=zbj_token)
        check("标记已读成功", mark_resp.get("code") == 200)
        
        # 验证已读后通知消失
        time.sleep(0.5)
        notif_resp2 = req("get", "/calendars/today-notifications", token=zbj_token)
        notifications2 = notif_resp2.get("data", [])
        still_has_notif = any(n.get("id") == notif_id for n in notifications2)
        check("已读通知从列表消失", not still_has_notif, f"剩余通知数={len(notifications2)}")
        
        # 验证刷新后也不会重新出现
        check("刷新后通知不重现", len(notifications2) < len(notifications))

    # 9. ylw也应该收到通知（创建者）
    print("\n=== 9. ylw检查自己的通知 ===")
    ylw_notif_resp = req("get", "/calendars/today-notifications", token=ylw_token)
    ylw_notifications = ylw_notif_resp.get("data", [])
    check("ylw也收到通知", len(ylw_notifications) > 0, f"通知数={len(ylw_notifications)}")

    # 10. ylw创建今天的日程（测试当天通知）
    print("\n=== 10. ylw创建今天的日程 ===")
    today_start = datetime.now().replace(hour=14, minute=0, second=0, microsecond=0)
    if today_start < datetime.now():
        today_start = datetime.now() + timedelta(minutes=30)
    today_end = today_start + timedelta(hours=1)
    
    today_resp = req("post", "/calendars", token=ylw_token, json={
        "title": "测试今天会议-ylw创建",
        "description": "今天的测试日程",
        "start_time": today_start.isoformat(),
        "end_time": today_end.isoformat(),
        "is_all_day": False,
        "participants": [
            {"user_dn": "zbj", "display_name": "zbj"}
        ]
    })
    check("ylw创建今天日程成功", today_resp.get("code") == 200)
    today_cal_id = today_resp.get("data", {}).get("id")

    # 11. zbj应该收到今天日程的通知
    print("\n=== 11. zbj检查今天日程通知 ===")
    time.sleep(1)
    zbj_notif_resp3 = req("get", "/calendars/today-notifications", token=zbj_token)
    zbj_notifications3 = zbj_notif_resp3.get("data", [])
    has_today_notif = any(n.get("calendar_id") == today_cal_id for n in zbj_notifications3)
    check("zbj收到今天日程通知", has_today_notif, f"总通知数={len(zbj_notifications3)}")

    # 12. 测试全部已读
    print("\n=== 12. 测试全部已读 ===")
    if zbj_notifications3:
        for n in zbj_notifications3:
            req("put", f"/calendars/notifications/{n['id']}/read", token=zbj_token)
        time.sleep(0.5)
        zbj_notif_resp4 = req("get", "/calendars/today-notifications", token=zbj_token)
        check("全部已读后通知列表为空", len(zbj_notif_resp4.get("data", [])) == 0)
        
        unread_resp2 = req("get", "/calendars/unread-count", token=zbj_token)
        check("全部已读后未读数为0", unread_resp2.get("data", {}).get("count", -1) == 0)

    # 汇总
    print("\n" + "=" * 50)
    passed = sum(1 for _, s, _ in results if s == "PASS")
    total = len(results)
    print(f"测试结果: {passed}/{total} 通过")
    if passed == total:
        print("🎉 所有测试通过！")
    else:
        print("\n失败项:")
        for name, status, detail in results:
            if status == "FAIL":
                print(f"  ❌ {name}: {detail}")
    
    return results

if __name__ == "__main__":
    test_all()
