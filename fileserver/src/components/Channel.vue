<template>
  <div class="card mt-2 mb-2">
    <div class="card-header">
      <h5 class="card-title">{{ channel }}</h5>
    </div>
    <div class="card-body">
      <div class="container">
        <div class="row">
          <div class="col">
            <h5 class="card-title">{{ clients }} Clients</h5>
          </div>
          <div class="col">
            <h5 class="card-title">{{ files }} Files sent</h5>
          </div>
        </div>
        <hr />
        <div class="row">
          <div class="col">
            <h6>Send a file to this channel</h6>
          </div>
        </div>

        <form
          @submit.prevent="submitHandler"
          autocomplete="off"
          method="POST"
          action="action"
          class="row row-cols-lg-auto g-3 align-items-center"
        >
          <div class="col-auto">
            <label class="visually-hidden" :for="`file-${channel}`">Name</label>
            <div class="input-group">
              <input
                class="form-control"
                type="file"
                :id="`file-${channel}`"
                :name="`file-${channel}`"
              />
            </div>
          </div>
          <div class="col-12">
            <button type="submit" class="btn btn-primary">Send</button>
          </div>
        </form>
      </div>
    </div>
    <div class="card-footer text-muted">
      Created at {{ new Date(created_at) }}
    </div>
  </div>
</template>

<script>
import notie from "notie"
export default {
  name: "Channel",
  props: ["channel", "files_sent","clients", "created_at", "event", "ref"],
  data() {
    return { files:this.files_sent,file: null, file_base64: null, file_name: null, file_extension:null}
  },
  methods: {
    submitHandler(e) {
      const elements = e.target.elements
      const fileInput = elements[`file-${this.channel}`]
      if (fileInput.files.length==1){
        notie.alert({
          type: 'success',
          text: `Sending ${fileInput.files[0].name} to ${this.channel}`,
        })
        //get file base 64
        var file = fileInput.files[0]
        const reader = new FileReader()
        reader.readAsDataURL(file)
        reader.onload = () => {
          this.file_base64 = reader.result.split(",")[1]
          this.file_extension = "."+file.name.split('.')[file.name.split('.').length-1]
          this.file_name = file.name.split('.').slice(0, -1).join('.')
          fileInput.value = ''
          //send file

          const payload = {
            channel:this.channel,
            name:this.file_name,
            extension: this.file_extension,
            base64:this.file_base64,

          }

          const requestOptions = {
            method:"POST",
            body: JSON.stringify(payload),
          }
          fetch("http://localhost:8081/channels/send",requestOptions)
          .then((response)=>response.json())
          .then((response)=>{
            if(response.error){
              notie.alert({
                type: "error",
                text: "Failed to send file",
              })
            }else{
              notie.alert({
                type: "success",
                text: "File sent",
              })
              this.files ++
            }
          })

        }
        reader.onerror = function (error) {
          notie.alert({
            type: "error",
            stay: true,
            text: "Unexpected error while reading file",
          })
          console.log("Error: ", error)
        }

      }else{
        notie.alert({
          type: 'error',
          text: 'Please select a maximun of one file to send',
        })
      }
    },
  },
}
</script>

<style>
</style>