<template>
  <div class="container">
    <div class="row">
      <div class="col">
        <h1>Channels</h1>

        <h4 v-if="items.length == 0">
          Please have a coffee while someone creates a channel on the server
          &#9749;
        </h4>

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
import Channel from "@/components/Channel.vue";
import {fetchWrapper} from "@/api/api.js";
import { sortByName } from "@/utils/helpers";
import { API_BASE_URL, CHANNELS_LIST_ENDPOINT } from "@/constants.js";

export default {
  name: "Channels",
  components: { Channel },
  props: [],
  data() {
    return { items: [] }
  },
  mounted() {
    let endpoint = `${API_BASE_URL}${CHANNELS_LIST_ENDPOINT}`
    fetchWrapper
      .get(endpoint)
      .then((data) => this.items = data.sort(sortByName))
      .catch((error) => {
        notie.alert({
            type: "error",
            stay: true,
            text: "There was an error when trying to get available channels",
          })
          console.log("Error: ", error)
      });
  },
};
</script>

<style>
</style>