from flask import (Blueprint, flash, g, redirect, render_template, request, session, url_for, make_response, get_flashed_messages)
import os
import json
import requests


admin_bp = Blueprint('admin', __name__, template_folder='templates',
                     static_folder='static')

peers=["peer0.org1.example.com","peer1.org1.example.com"]


#主页
@admin_bp.route('/')
def index():
    if request.cookies.get("name"):
        return render_template("index.html")
    else:
        return render_template("login.html")


#登录页面
@admin_bp.route('/login', methods=['POST', 'GET'])
def login():
    print('login')
    if request.method == 'GET':
        return redirect("admin")
    else:
        #print(request.form)
        user = request.form['username']
        pwd = request.form['password']
        org = request.form['orgName']
        res=requests.post("http://localhost:4000/users",
                          "username=%s&orgName=%s&password=%s"%(user,org,pwd),
                         headers={"content-type": "application/x-www-form-urlencoded"})
        if res.status_code != 200:
            return render_template("error.html",message="status_code: "+res.status_code+res.text)
        restext=json.loads(res.text)
        print(restext)
        if restext['success']!=True:
            return render_template("error.html",message=restext['message'])

        resp = make_response(redirect("admin"))
        resp.set_cookie('name',user)
        resp.set_cookie('token', restext['token'])
        return resp


#注册页
@admin_bp.route('/signup')
def signup():
    return render_template("sign-up.html")


#错误页面
@admin_bp.route('/error/<message>')
def error(message):
    return render_template("error.html",message=message)


#信息页面
@admin_bp.route('/info/<message>')
def info(message):
    return render_template("info.html",message=message)


#进货页面
@admin_bp.route('/purchase')
def purchase():
    return render_template("purchase.html")


'''表单处理'''
#处理注册
@admin_bp.route('/register', methods=['POST'])
def register():
    user = request.form['username']
    pwd = request.form['password']
    org = request.form['orgName']
    res = requests.post("http://localhost:4000/register",
                        "username=%s&orgName=%s&password=%s" % (user, org, pwd),
                        headers={"content-type": "application/x-www-form-urlencoded"})
    if res.status_code != 200:
        return render_template("error.html", message="status_code: " + res.status_code + res.text)
    restext = json.loads(res.text)
    print(restext)
    if restext['success'] != True:
        return render_template("error.html", message=restext['message'])

    resp = make_response(redirect("admin"))
    resp.set_cookie('name', user)
    resp.set_cookie('token', restext['token'])
    return resp


#处理退出
@admin_bp.route('/logout', methods=['GET'])
def logout():
    resp = make_response(redirect("admin"))
    resp.delete_cookie('name')
    resp.delete_cookie('token')
    return resp


#查询商品处理
@admin_bp.route('/query_category', methods=['POST'])
def query_category():
    print(request.form)
    #参数
    token = request.cookies.get('token')
    headers = {"authorization": "Bearer "+token, "content-type": "application/json"}
    channel_name="mychannel"
    chaincode_name="category"
    data = {
        "peers": peers,
        "fcn": "query",
        "args": [request.form['id'], request.form['store_id']]
    }
    #post
    res=requests.post("http://localhost:4000/channels/%s/chaincodes/%s"%(channel_name,chaincode_name),data=json.dumps(data),headers=headers)
    if res.status_code != 200:
        return render_template("error.html", message="status_code: " + res.status_code + res.text)
    print(res.text)
    try:
        restext = json.loads(res.text)
        print(restext)
        if restext['success'] != True:
            return render_template("error.html", message=restext['message'])
        return redirect(url_for("admin.info", message=restext['message']))
    except:
        return render_template("error.html", message=res.text)


# 查询商品处理
@admin_bp.route('/insert_category', methods=['POST'])
def insert_category():
    print(request.form)
    # 参数
    token = request.cookies.get('token')
    headers = {"authorization": "Bearer " + token, "content-type": "application/json"}
    channel_name = "mychannel"
    chaincode_name = "category"
    data = {
        "peers": peers,
        "fcn": "insert",
        "args": [json.dumps(request.form)]
    }
    # post
    res = requests.post("http://localhost:4000/channels/%s/chaincodes/%s" % (channel_name, chaincode_name), data=json.dumps(data),
                        headers=headers)
    if res.status_code != 200:
        return render_template("error.html", message="status_code: " + res.status_code + res.text)

    try:
        restext = json.loads(res.text)
        print(restext)
        if restext['success'] != True:
            return render_template("error.html", message=restext['message'])
        return redirect(url_for("admin.info", message=restext['message']))
    except:
        return render_template("error.html", message=res.text)
