<template>
  <div class="container">
    <div class="row">
      <div class="col">
        <h1>FileServer Stats</h1>
        <div v-if="stats.createdAt !== undefined">
          <stat
            :header="'Files Sent'"
            :badge="stats.filesSent"
            :prefix="'Clients have sent'"
            :suffix="'files.'"
          ></stat>
          <stat
            :header="'Clients Connected'"
            :badge="stats.clientsConnected"
            :prefix="'There are'"
            :suffix="'clients connected.'"
          ></stat>
          <stat
            :header="'Channels'"
            :badge="stats.channelsAvailable"
            :prefix="'There are'"
            :suffix="'channels available.'"
          ></stat>
          <stat-time-since
            :header="'Uptime'"
            :date="stats.createdAt"
            :prefix="'Server has been running for'"
            :suffix="''"
          ></stat-time-since>
        </div>
        <div v-else>Oh no, seems like the server is down 😢</div>
      </div>
    </div>
  </div>
</template>

<script>
import Stat from "@/components/Stat.vue";
import StatTimeSince from "@/components/StatTimeSince.vue";
import { fetchWrapper } from "@/api/api.js";
import { API_BASE_URL, STATS_ENDPOINT } from "@/constants.js";
export default {
  name: "Body",
  components: { Stat, StatTimeSince },
  data() {
    return { stats: {} };
  },
  mounted() {
    let endpoint = `${API_BASE_URL}${STATS_ENDPOINT}`;

    fetchWrapper
      .get(endpoint)
      .then(
        (data) =>
          (this.stats = {
            filesSent: data.files_sent,
            clientsConnected: data.clients_connected,
            channelsAvailable: data.channels_available,
            createdAt: new Date(data.created_at),
          })
      )
      .catch((error) => {
        notie.alert({
          type: "error",
          stay: true,
          text: "There was an error when trying to get available channels",
        });
        console.log("Error: ", error);
      });
  },
};
</script>

<style>
</style>