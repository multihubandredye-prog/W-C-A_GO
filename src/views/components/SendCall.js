// export Vue Component
import FormRecipient from "./generic/FormRecipient.js";

export default {
    name: 'SendCall',
    components: {
        FormRecipient
    },
    data() {
        return {
            phone: '',
            type: window.TYPEUSER,
            loading: false,
            duration: 15,
            audio_file: null,
            audio_url: '',
            audio_path: '',
        }
    },
    computed: {
        phone_id() {
            return this.phone + this.type;
        },
        hasAudioSource() {
            return !!(this.audio_file || this.audio_url.trim() || this.audio_path.trim());
        }
    },
    methods: {
        openModal() {
            $('#modalSendCall').modal({
                onApprove: function () {
                    return false;
                }
            }).modal('show');
        },
        onFileChange(event) {
            const file = event.target.files && event.target.files[0];
            if (file) {
                // The audio sources are mutually exclusive.
                this.audio_url = '';
                this.audio_path = '';
                this.audio_file = file;
            } else {
                this.audio_file = null;
            }
        },
        clearAudioFile() {
            this.audio_file = null;
            if (this.$refs.audioFileInput) {
                this.$refs.audioFileInput.value = '';
            }
        },
        isValidForm() {
            if (this.type !== window.TYPESTATUS && !this.phone.trim()) {
                return false;
            }
            if (this.duration < 1 || this.duration > 3600) {
                return false;
            }
            return true;
        },
        async handleSubmit() {
            if (!this.isValidForm() || this.loading) {
                return;
            }
            try {
                let response = await this.submitApi()
                window.showSuccessInfo(response)
                $('#modalSendCall').modal('hide');
            } catch (err) {
                window.showErrorInfo(err)
            }
        },
        async submitApi() {
            this.loading = true;
            try {
                if (this.audio_file) {
                    // Real file upload: multipart/form-data.
                    const form = new FormData();
                    form.append('phone', this.phone_id);
                    form.append('duration', this.duration);
                    form.append('audio', this.audio_file);

                    const response = await window.http.post(`/send/call`, form)
                    this.handleReset();
                    return response.data.message;
                }

                const payload = {
                    phone: this.phone_id,
                    duration: this.duration
                }
                if (this.audio_url.trim()) payload.audio_url = this.audio_url;
                if (this.audio_path.trim()) payload.audio_path = this.audio_path;

                const response = await window.http.post(`/send/call`, payload)
                this.handleReset();
                return response.data.message;
            } catch (error) {
                if (error.response) {
                    throw new Error(error.response.data.message);
                }
                throw new Error(error.message);
            } finally {
                this.loading = false;
            }
        },
        handleReset() {
            this.phone = '';
            this.type = window.TYPEUSER;
            this.duration = 15;
            this.clearAudioFile();
            this.audio_url = '';
            this.audio_path = '';
        }
    },
    template: `
    <div class="green card" @click="openModal()" style="cursor: pointer">
        <div class="content">
            <a class="ui green right ribbon label">Send</a>
            <div class="header">Send Call</div>
            <div class="description">
                Place a VoIP call; audio plays when answered
            </div>
        </div>
    </div>

    <!--  Modal SendCall  -->
    <div class="ui small modal" id="modalSendCall">
        <i class="close icon"></i>
        <div class="header">
            Send Call
        </div>
        <div class="content">
            <form class="ui form">
                <FormRecipient v-model:type="type" v-model:phone="phone"/>

                <div class="field">
                    <label>Max Duration (seconds)</label>
                    <input v-model.number="duration" type="number" min="1" max="3600"
                           placeholder="15" aria-label="call duration">
                    <div class="ui pointing label">
                        Maximum call time. With audio, the call also ends when the file finishes.
                    </div>
                </div>

                <div class="field">
                    <label>Audio File (MP3)</label>
                    <div class="ui action input">
                        <input ref="audioFileInput" type="file" accept="audio/mpeg,.mp3"
                               @change="onFileChange" aria-label="call audio file">
                        <button class="ui icon button" type="button" title="Clear file"
                                v-if="audio_file" @click="clearAudioFile">
                            <i class="trash icon"></i>
                        </button>
                    </div>
                    <div class="ui pointing label">
                        Upload the audio file to play when the recipient answers.
                        Only one audio source: the file, a URL or Base64.
                    </div>
                </div>

                <div class="field" :class="{'disabled': !!audio_file}">
                    <label>Audio URL (optional, MP3)</label>
                    <input v-model="audio_url" type="text" placeholder="https://meusite.com/audio/alerta.mp3"
                           aria-label="call audio url" :disabled="!!audio_file">
                </div>
                <div class="field" :class="{'disabled': !!audio_file}">
                    <label>Audio Base64 (optional, MP3)</label>
                    <textarea v-model="audio_path" rows="2"
                              placeholder="data:audio/mp3;base64,... or a raw base64 string"
                              aria-label="call audio base64" :disabled="!!audio_file"></textarea>
                </div>
            </form>
        </div>
        <div class="actions">
            <button class="ui approve positive right labeled icon button" :class="{'loading': this.loading, 'disabled': !isValidForm() || loading}"
                 @click.prevent="handleSubmit">
                Call
                <i class="phone icon"></i>
            </button>
        </div>
    </div>
`
}
