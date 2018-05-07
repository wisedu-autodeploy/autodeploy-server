webpackJsonp([1],{"+skl":function(t,e){},JFzA:function(t,e){},NHnr:function(t,e,a){"use strict";Object.defineProperty(e,"__esModule",{value:!0});var n=a("7+uW"),r=a("mtWM"),s=a.n(r),p={render:function(){var t=this.$createElement,e=this._self._c||t;return e("div",{attrs:{id:"app"}},[e("router-view")],1)},staticRenderFns:[]};var i=a("VU/8")({created:function(){}},p,!1,function(t){a("P5/u")},null,null).exports,o=a("/ocq"),l=a("Xxa5"),u=a.n(l),c=a("exGp"),d=a.n(c),m=a("Y4FN"),h=a.n(m),v={name:"HelloWorld",data:function(){return{username:"",password:"",gitlabAppIndex:null,marathonAppIndex:null,gitlabApps:[],marathonApps:[],currentStep:0,logining:!1}},computed:{hasLogin:function(){return this.gitlabApps.length>0||this.marathonApps.length>0},selectedGitlabApp:function(){return null===this.gitlabAppIndex?null:this.gitlabApps[this.gitlabAppIndex]},selectedMarathonApp:function(){return null===this.marathonAppIndex?null:this.marathonApps[this.marathonAppIndex]}},methods:{login:function(){var t=this;return d()(u.a.mark(function e(){var a,n,r,p,i;return u.a.wrap(function(e){for(;;)switch(e.prev=e.next){case 0:return h.a.set("userInfo",{username:t.username,password:t.password}),t.logining=!0,e.next=4,s.a.post("/v1/api/login",{username:t.username,password:t.password});case 4:if(a=e.sent,n=a.data,t.logining=!1,r=n.code,p=n.data,i=n.message,"0"===r){e.next=11;break}return t.$Message.error({content:i}),e.abrupt("return");case 11:h.a.set("appData",p),t.gitlabApps=p.gitlabApps||t.gitlabApps,t.marathonApps=p.marathonApps||t.marathonApps,t.currentStep=1;case 15:case"end":return e.stop()}},e,t)}))()},deploy:function(){var t=this;return d()(u.a.mark(function e(){var a,n,r,p;return u.a.wrap(function(e){for(;;)switch(e.prev=e.next){case 0:return e.next=2,s.a.post("/v1/api/autodeploy",{username:t.username,password:t.password,maintainer:t.selectedGitlabApp.Maintainer,name:t.selectedGitlabApp.Name,marathonName:t.selectedMarathonApp});case 2:if(a=e.sent,n=a.data,r=n.code,n.data,p=n.message,"0"===r){e.next=8;break}return t.$Message.error({content:p}),e.abrupt("return");case 8:case"end":return e.stop()}},e,t)}))()},prevStep:function(){this.currentStep-=1},nextStep:function(){this.currentStep+=1}},created:function(){var t=h.a.get("userInfo",{}),e=h.a.get("appData",null);this.username=t.username,this.password=t.password,e?(this.gitlabApps=e.gitlabApps,this.marathonApps=e.marathonApps,this.currentStep=1):this.username&&this.password?this.login():this.currentStep=0}},f={render:function(){var t=this,e=t.$createElement,a=t._self._c||e;return a("div",{staticClass:"hello"},[a("div",{staticStyle:{"margin-bottom":"16px"}},[a("Steps",{attrs:{current:t.currentStep}},[a("Step",{attrs:{title:"登录",content:"登录 gitlab"}}),t._v(" "),a("Step",{attrs:{title:"部署",content:"选择项目部署"}})],1)],1),t._v(" "),a("div",{staticStyle:{display:"flex","flex-direction":"column","align-items":"center"}},[a("div",{directives:[{name:"show",rawName:"v-show",value:0===t.currentStep,expression:"currentStep === 0"}],staticStyle:{width:"400px"}},[a("Input",{attrs:{placeholder:"username"},model:{value:t.username,callback:function(e){t.username=e},expression:"username"}}),t._v(" "),a("Input",{attrs:{placeholder:"password",type:"password"},model:{value:t.password,callback:function(e){t.password=e},expression:"password"}}),t._v(" "),a("div",{staticStyle:{"margin-top":"8px"}},[a("Button",{attrs:{type:"default",disabled:!t.hasLogin},on:{click:t.nextStep}},[t._v("下一步")]),t._v(" "),a("Button",{attrs:{type:"primary"},on:{click:t.login}},[t._v("登录")])],1)],1),t._v(" "),a("div",{directives:[{name:"show",rawName:"v-show",value:1===t.currentStep,expression:"currentStep === 1"}],staticStyle:{width:"400px"}},[a("Select",{attrs:{placeholder:"选择 gitlab 项目",filterable:""},model:{value:t.gitlabAppIndex,callback:function(e){t.gitlabAppIndex=e},expression:"gitlabAppIndex"}},t._l(t.gitlabApps,function(e,n){return a("Option",{key:e.Maintainer+"/"+e.Name,attrs:{value:n}},[t._v("\n            "+t._s(e.Maintainer+"/"+e.Name)+"\n          ")])})),t._v(" "),a("Select",{attrs:{placeholder:"选择 marathon 项目",filterable:""},model:{value:t.marathonAppIndex,callback:function(e){t.marathonAppIndex=e},expression:"marathonAppIndex"}},t._l(t.marathonApps,function(e,n){return a("Option",{key:e,attrs:{value:n}},[t._v("\n            "+t._s(e)+"\n          ")])})),t._v(" "),a("div",{staticStyle:{"margin-top":"8px"}},[a("Button",{attrs:{type:"default"},on:{click:t.prevStep}},[t._v("上一步")]),t._v(" "),a("Button",{attrs:{type:"primary"},on:{click:t.deploy}},[t._v("部署")])],1)],1)]),t._v(" "),t.logining?a("Spin",{attrs:{size:"large",fix:""}}):t._e()],1)},staticRenderFns:[]};var g=a("VU/8")(v,f,!1,function(t){a("JFzA")},"data-v-64211144",null).exports;n.default.use(o.a);var A=new o.a({routes:[{path:"/",name:"HelloWorld",component:g}]}),x=a("BTaQ"),b=a.n(x);a("+skl");n.default.config.productionTip=!1,n.default.use(b.a),new n.default({el:"#app",router:A,components:{App:i},template:"<App/>"})},"P5/u":function(t,e){}},["NHnr"]);
//# sourceMappingURL=app.6fa94df1bad8e88c32c8.js.map