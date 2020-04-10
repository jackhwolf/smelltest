import requests
import json
import time 

def decode_go_resp(resp):
    resp = resp.text
    resp = json.loads(resp)
    print(resp)
    print()
    return resp

base_url = "http://localhost:8081/v1/{}/"
user_url = base_url.format('user')
smell_url = base_url.format('smells')


def test_signup_user(**kw):
    user = {
        "username": "skippyelvis",
        "password": "password",
    }
    user.update(*kw)
    resp = requests.post(user_url, data=json.dumps(user))
    print("Signup user")
    resp = decode_go_resp(resp)
    return resp

def test_login_user(**kw):
    user = {
        "username": "skippyelvis",
        "password": "password",
    }
    user.update(*kw)
    resp = requests.post(user_url + 'login/', data=json.dumps(user))
    print("Login user")
    resp = decode_go_resp(resp)
    return resp

def test_logout_user(token, **kw):
    user = {
        "X-Access-Token": token,
    }
    user.update(*kw)
    resp = requests.get(user_url + 'logout/', headers=user)
    print("Logout user")
    resp = decode_go_resp(resp)
    return resp

def test_delete_user(**kw):
    user = {
        "username": "skippyelvis",
        "password": "password",
    }
    user.update(*kw)
    resp = requests.delete(user_url, data=json.dumps(user))
    print("Delete user")
    resp = decode_go_resp(resp)
    return resp

def test_get_smells(token, **kw):
    headers = {
        "X-Access-Token": token,
    }
    headers.update(*kw)
    resp = requests.get(smell_url, headers=headers)
    print("Get smells")
    resp = decode_go_resp(resp)
    return resp

if __name__ == "__main__":
    r1 = test_signup_user()
    r2 = test_login_user()
    test_get_smells(r1['Token'])
    test_logout_user(r1['Token'])
    test_logout_user(r2['Token'])
    test_delete_user()
