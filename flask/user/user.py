from flask import Blueprint
from flask import (Blueprint, flash, g, redirect, render_template, request, session, url_for, make_response)
import requests
import json
import time

user_bp = Blueprint('user', __name__)

orgName='store1'
peers=["peer1.store1.aliyunbaas.com:31111"]
channel_name="first-channel"
user_cc='usercc'
category_cc='category'
commodity_cc='commodity'



'''页面'''
#首页
@user_bp.route('/')
def index():
    if not request.cookies.get("username"):
        return render_template("user/login.html")

    token = request.cookies.get('token')
    print(token)
    headers = {"authorization": "Bearer " + token, "content-type": "application/json"}

    data = {
        "peers": peers,
        "fcn": "queryByID",
        "args": [request.cookies.get('username')]
    }
    # post
    try:
        res = requests.post("http://localhost:4000/channels/%s/chaincodes/%s" % (channel_name, user_cc),
                             data=json.dumps(data), headers=headers)
    except Exception as e:
        return render_template("user/error.html", message=e)

    if res.status_code != 200:
        return render_template("user/error.html", message="status_code: " + str(res.status_code) + res.text)

    try:
        restext = json.loads(res.text)
        print(restext)
        if restext['success'] != True:
            return render_template("user/error.html", message=restext['message'])
        # return redirect(url_for("user.info", message=res2text['message']))
    except:
        return render_template("user/error.html", message=res.text)

    return render_template("user/index.html",user_info=json.loads(restext['message']))


#注册页
@user_bp.route('/signup')
def signup():
    return render_template("user/sign-up.html")


#错误页面
@user_bp.route('/error/<message>')
def error(message):
    return render_template("user/user/error.html",message=message)


#信息页面
@user_bp.route('/info/<message>')
def info(message):
    return render_template("user/info.html",message=message)


'''表单处理'''
#处理注册
@user_bp.route('/register', methods=['POST'])
def register():
    print(request.form)
    user = request.form['ID']
    pwd = request.form['Password']
    org = orgName
    res = requests.post("http://localhost:4000/users",
                        "username=%s&orgName=%s&password=%s" % (user, org, pwd),
                        headers={"content-type": "application/x-www-form-urlencoded"})
    if res.status_code != 200:
        return render_template("user/error.html", message="status_code: " + str(res.status_code) + res.text)
    restext = json.loads(res.text)
    print(restext)
    if restext['success'] != True:
        return render_template("user/error.html", message=restext['message'])

    #写入用户信息
    print(request.form)
    token=restext['token']
    print(token)
    headers = {"authorization": "Bearer " + token, "content-type": "application/json"}

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
        return render_template("user/error.html",message=e)

    if res2.status_code != 200:
        return render_template("user/error.html", message="status_code: " + str(res2.status_code) + res2.text)

    try:
        res2text = json.loads(res2.text)
        print(res2text)
        if res2text['success'] != True:
            return render_template("user/error.html", message=res2text['message'])
        #return redirect(url_for("user.info", message=res2text['message']))
    except:
        return render_template("user/error.html", message=res2.text)

    resp = make_response(redirect("user"))
    resp.set_cookie('username', user, max_age=3600)
    resp.set_cookie('token', restext['token'], max_age=3600)
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
            return render_template("user/error.html",message="status_code: "+str(res.status_code)+res.text)
        restext=json.loads(res.text)
        print(restext)
        if restext['success']!=True:
            return render_template("user/error.html",message=restext['message'])

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


#处理查询商品
@user_bp.route('/query', methods=['POST'])
def query_commodity():
    print(request.form)
    # 参数
    token = request.cookies.get('token')
    headers = {"authorization": "Bearer " + token, "content-type": "application/json"}

    #查询commodity
    data = {
        "peers": peers,
        "fcn": "queryByID",
        "args": [request.form['ID']]
    }
    # post
    try:
        res = requests.post("http://localhost:4000/channels/%s/chaincodes/%s" % (channel_name, commodity_cc), data=json.dumps(data),
                        headers=headers)
    except Exception as e:
        return render_template("user/error.html", message=e)
    if res.status_code != 200:
        return render_template("user/error.html", message="status_code: " + str(res.status_code) + res.text)
    try:
        restext = json.loads(res.text)
        print(restext)
        if restext['success'] != True:
            return render_template("user/error.html", message=restext['message'])
        #return redirect(url_for("admin.info", message=restext['message']))
    except:
        return render_template("user/error.html", message=res.text)
    category=restext['message']['Category']
    storeID=restext['message']['StoreID']
    # 查询category
    data = {
        "peers": peers,
        "fcn": "query",
        "args": [category,storeID]
    }
    # post
    try:
        res2 = requests.post("http://localhost:4000/channels/%s/chaincodes/%s" % (channel_name, category_cc),
                            data=json.dumps(data),
                            headers=headers)
    except Exception as e:
        return render_template("user/error.html", message=e)
    if res2.status_code != 200:
        return render_template("user/error.html", message="status_code: " + str(res2.status_code) + res2.text)

    try:
        res2text = json.loads(res2.text)
        print(res2text)
        if res2text['success'] != True:
            return render_template("user/error.html", message=res2text['message'])
        return redirect(url_for("user.info", message=restext['message']+'\n'+res2text['message']))
    except:
        return render_template("user/error.html", message=res.text+'\n'+res2.text)


#查询商品处理
@user_bp.route('/query_category', methods=['POST'])
def query_category():
    print(request.form)
    #参数
    token = request.cookies.get('token')
    headers = {"authorization": "Bearer "+token, "content-type": "application/json"}

    data = {
        "peers": peers,
        "fcn": "query",
        "args": [request.form['id'], request.form['store_id']]
    }
    #post
    res=requests.post("http://localhost:4000/channels/%s/chaincodes/%s"%(channel_name,category_cc),data=json.dumps(data),headers=headers)

    if res.status_code != 200:
        return res.text
        #return render_template("user/error.html", message="status_code: " + str(res.status_code) + res.text)
    print(res.text)
    try:
        restext = json.loads(res.text)
        print(restext)
        if restext['success'] != True:
            return render_template("user/error.html", message=restext['message'])
        return redirect(url_for("user.info", message=restext['message']))
    except:
        return render_template("user/error.html", message=res.text)