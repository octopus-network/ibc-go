(window.webpackJsonp=window.webpackJsonp||[]).push([[18],{558:function(e,t,a){},592:function(e,t,a){"use strict";a(558)},615:function(e,t,a){"use strict";a.r(t);a(19),a(43);var n={data:function(){return{language:null,languageList:null}},created:function(){this.language=this.$localeConfig.path,this.languageList=this.$site.locales},methods:{select:function(e){this.$router.push(this.$page.path.replace(this.$localeConfig.path,e.target.value))}}},l=(a(592),a(1)),u=Object(l.a)(n,(function(){var e=this,t=e.$createElement,a=e._self._c||t;return a("div",[a("div",{staticClass:"select"},[a("select",{directives:[{name:"model",rawName:"v-model",value:e.language,expression:"language"}],on:{input:e.select,change:function(t){var a=Array.prototype.filter.call(t.target.options,(function(e){return e.selected})).map((function(e){return"_value"in e?e._value:e.value}));e.language=t.target.multiple?a:a[0]}}},e._l(e.languageList,(function(t,n){return a("option",{domProps:{value:n}},[e._v(e._s(t.label||t.lang))])})),0)])])}),[],!1,null,"5404a90a",null);t.default=u.exports}}]);