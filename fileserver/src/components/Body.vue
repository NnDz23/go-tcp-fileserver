<template>
  <div class="container">
    <div class="row">
      <div class="col">
        <h1>FileServer Stats</h1>
        <div v-if="stats.createdAt !==undefined">
          <stat :header="'Files Sent'" :badge="stats.filesSent" :prefix="'Clients have sent'" :suffix="'files.'"></stat>
          <stat :header="'Clients Connected'" :badge="stats.clientsConnected" :prefix="'There are'" :suffix="'clients connected.'"></stat>
          <stat :header="'Channels'" :badge="stats.channelsAvailable" :prefix="'There are'" :suffix="'channels available.'"></stat>
          <stat-time-since :header="'Uptime'" :date="stats.createdAt" :prefix="'Server has been running for'" :suffix="''"></stat-time-since>
        </div>
        <div v-else>Oh no, seems like the server is down 😢</div>
      </div>
    </div>
  </div>
</template>

<script>
import Stat from "./Stat.vue"
import StatTimeSince from './StatTimeSince.vue'
export default {
  name: "Body",
  components: {Stat, StatTimeSince},
  data (){
    return {stats:{}}
  },
  mounted(){
    fetch('http://localhost:8081/stats')
    .then((response) => response.json())
    .then((data) => {
      this.stats = {
        filesSent: data.files_sent,
        clientsConnected: data.clients_connected,
        channelsAvailable : data.channels_available,
        createdAt: new Date(data.created_at)
      }
    })

    
  }
}
</script>

<style>

</style>