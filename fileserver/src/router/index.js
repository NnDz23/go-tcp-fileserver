import {createRouter, createWebHistory} from 'vue-router'
import Body from './../components/Body.vue'
import Channels from './../components/Channels.vue'

const routes = [
    {
        path:"/",
        name:"Home",
        component:Body
    },
    {
        path:"/channels",
        name:"Channels",
        component:Channels
    }
]

const router = createRouter({history:createWebHistory(),routes})
export default router