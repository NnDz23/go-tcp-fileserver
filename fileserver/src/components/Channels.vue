<template>
  <div class="container">
    <div class="row">
      <div class="col">
        <h1>Channels</h1>

        <h4 v-if="items.length == 0">Please have a coffe while somebody creates a channel on the server &#9749;</h4>

        <channel
          v-for="item in items"
          :key="item.name"
          :channel="item.name"
          :clients="item.clients_connected"
          :files_sent="item.files_sent"
          :created_at="item.created_at"
        ></channel>
      </div>
    </div>
  </div>
</template>

<script>
import Channel from "./Channel.vue";
export default {
  name: "Channels",
  components:{Channel},
  props: [],
  data (){
    return {items:[]}
  },
  mounted(){
    
    fetch('http://localhost:8081/channels/list')
    .then((response) => response.json())
    .then((data) => this.items = data.sort((a,b) => (a.name > b.name) ? 1 : ((b.name > a.name) ? -1 : 0)));
  }
};
</script>

<style>
</style>