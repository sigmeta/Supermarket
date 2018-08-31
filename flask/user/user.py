from flask import Blueprint
from flask import (Blueprint, flash, g, redirect, render_template, request, session, url_for, make_response)
import requests
import json


user_bp = Blueprint('user', __name__)

orgName='Org1'
peers=["peer0.org1.example.com","peer1.org1.example.com"]
channel_name="mychannel"
user_cc='users'



'''页面'''
#首页
@user_bp.route('/')
def index():
    if not request.cookies.get("username"):
        return render_template("user/login.html")
    return render_template("user/index.html")


#注册页
@user_bp.route('/signup')
def signup():
    return render_template("user/sign-up.html")


#错误页面
@user_bp.route('/error/<message>')
def error(message):
    return render_template("user/error.html",message=message)


#信息页面
@user_bp.route('/info/<message>')
def info(message):
    return render_template("user/info.html",message=message)


'''表单处理'''
#处理注册
@user_bp.route('/register', methods=['POST'])
def register():
    user = request.form['ID']
    pwd = request.form['Password']
    res = requests.post("http://localhost:4000/register",
                        "username=%s&orgName=%s&password=%s" % (user, orgName, pwd),
                        headers={"content-type": "application/x-www-form-urlencoded"})
    if res.status_code != 200:
        return render_template("error.html", message="status_code: " + str(res.status_code) + res.text)
    restext = json.loads(res.text)
    print(restext)
    if restext['success'] != True:
        return render_template("error.html", message=restext['message'])
    resp = make_response(redirect("user"))
    resp.set_cookie('username', user, max_age=3600)
    resp.set_cookie('token', restext['token'], max_age=3600)

    #写入用户信息

    headers = {"authorization": "Bearer " + restext['token'], "content-type": "application/json"}

    data = {
        "peers": peers,
        "fcn": "insert",
        "args": [json.dumps(request.form)]
    }
    # post
    try:
        res2 = requests.post("http://localhost:4000/channels/%s/chaincodes/%s" % (channel_name, user_cc),
                        data=json.dumps(data), headers=headers)
    except Exception as e:
        return render_template("error.html",message=e)

    if res2.status_code != 200:
        return render_template("error.html", message="status_code: " + str(res2.status_code) + res2.text)

    try:
        res2text = json.loads(res2.text)
        print(res2text)
        if res2text['success'] != True:
            return render_template("error.html", message=res2text['message'])
        #return redirect(url_for("user.info", message=res2text['message']))
    except:
        return render_template("error.html", message=res2.text)

    return resp


#处理登录
@user_bp.route('/login', methods=['POST', 'GET'])
def login():
    print('login')
    if request.method == 'GET':
        return redirect("user")
    else:
        #print(request.form)
        user = request.form['username']
        pwd = request.form['password']
        org = orgName
        res=requests.post("http://localhost:4000/users",
                          "username=%s&orgName=%s&password=%s"%(user,org,pwd),
                         headers={"content-type": "application/x-www-form-urlencoded"})
        if res.status_code != 200:
            return render_template("error.html",message="status_code: "+str(res.status_code)+res.text)
        restext=json.loads(res.text)
        print(restext)
        if restext['success']!=True:
            return render_template("error.html",message=restext['message'])

        resp = make_response(redirect("user"))
        resp.set_cookie('username',user, max_age=3600)
        resp.set_cookie('token', restext['token'], max_age=3600)
        return resp


#处理退出
@user_bp.route('/logout', methods=['GET'])
def logout():
    resp = make_response(redirect("user"))
    resp.delete_cookie('username')
    resp.delete_cookie('token')
    return resp




