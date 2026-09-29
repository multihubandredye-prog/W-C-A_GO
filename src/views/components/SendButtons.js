// export Vue Component
import FormRecipient from "./generic/FormRecipient.js";

const BUTTON_TYPES = [
    { value: 'reply', text: 'Reply (quick answer)' },
    { value: 'cta_url', text: 'URL (open link)' },
    { value: 'cta_call', text: 'Call (dial number)' },
    { value: 'copy', text: 'Copy (copy code)' },
];

export default {
    name: 'SendButtons',
    components: {
        FormRecipient
    },
    data() {
        return {
            phone: '',
            type: window.TYPEUSER,
            loading: false,
            body: '',
            title: '',
            footer: '',
            image_url: '',
            buttons: [this.newButton(), this.newButton()],
        }
    },
    computed: {
        phone_id() {
            return this.phone + this.type;
        },
        buttonTypes() {
            return BUTTON_TYPES;
        }
    },
    methods: {
        newButton() {
            return { type: 'reply', title: '', id: '', url: '', phone_number: '', copy_code: '' };
        },
        openModal() {
            $('#modalSendButtons').modal({
                onApprove: function () {
                    return false;
                }
            }).modal('show');
        },
        isButtonTypeValid(button) {
            if (!button.title.trim()) {
                return false;
            }
            switch (button.type) {
                case 'cta_url':
                    return button.url.trim().length > 0;
                case 'cta_call':
                    return button.phone_number.trim().length > 0;
                case 'copy':
                    return button.copy_code.trim().length > 0;
                default:
                    return true;
            }
        },
        isValidForm() {
            if (this.type !== window.TYPESTATUS && !this.phone.trim()) {
                return false;
            }
            if (!this.body.trim()) {
                return false;
            }
            if (this.buttons.length < 1 || this.buttons.length > 3) {
                return false;
            }
            return this.buttons.every(button => this.isButtonTypeValid(button));
        },
        async handleSubmit() {
            if (!this.isValidForm() || this.loading) {
                return;
            }
            try {
                let response = await this.submitApi()
                window.showSuccessInfo(response)
                $('#modalSendButtons').modal('hide');
            } catch (err) {
                window.showErrorInfo(err)
            }
        },
        async submitApi() {
            this.loading = true;
            try {
                const payload = {
                    phone: this.phone_id,
                    body: this.body,
                    buttons: this.buttons.map(button => {
                        const item = { type: button.type, title: button.title };
                        if (button.type === 'reply' && button.id.trim()) {
                            item.id = button.id;
                        }
                        if (button.type === 'cta_url') {
                            item.url = button.url;
                        }
                        if (button.type === 'cta_call') {
                            item.phone_number = button.phone_number;
                        }
                        if (button.type === 'copy') {
                            item.copy_code = button.copy_code;
                        }
                        return item;
                    })
                }
                if (this.title.trim()) payload.title = this.title;
                if (this.footer.trim()) payload.footer = this.footer;
                if (this.image_url.trim()) payload.image_url = this.image_url;

                const response = await window.http.post(`/send/buttons`, payload)
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
            this.body = '';
            this.title = '';
            this.footer = '';
            this.image_url = '';
            this.buttons = [this.newButton(), this.newButton()];
        },
        addButton() {
            if (this.buttons.length < 3) {
                this.buttons.push(this.newButton())
            }
        },
        deleteButton(index) {
            this.buttons.splice(index, 1)
        }
    },
    template: `
    <div class="purple card" @click="openModal()" style="cursor: pointer">
        <div class="content">
            <a class="ui purple right ribbon label">Send</a>
            <div class="header">Send Buttons</div>
            <div class="description">
                Send interactive buttons (quick reply, URL, call, copy)
            </div>
        </div>
    </div>

    <!--  Modal SendButtons  -->
    <div class="ui small modal" id="modalSendButtons">
        <i class="close icon"></i>
        <div class="header">
            Send Buttons
        </div>
        <div class="content">
            <form class="ui form">
                <FormRecipient v-model:type="type" v-model:phone="phone"/>

                <div class="field">
                    <label>Body</label>
                    <textarea v-model="body" rows="2" placeholder="Main message text"
                              aria-label="buttons body"></textarea>
                </div>
                <div class="two fields">
                    <div class="field">
                        <label>Title (optional)</label>
                        <input v-model="title" type="text" placeholder="Header above the body"
                               aria-label="buttons title">
                    </div>
                    <div class="field">
                        <label>Footer (optional)</label>
                        <input v-model="footer" type="text" placeholder="Small text below the buttons"
                               aria-label="buttons footer">
                    </div>
                </div>
                <div class="field">
                    <label>Header Image URL (optional)</label>
                    <input v-model="image_url" type="text"
                           placeholder="https://... or data:image/...;base64,..."
                           aria-label="buttons header image">
                </div>

                <div class="field">
                    <label>Buttons (1 to 3)</label>
                    <div style="display: flex; flex-direction: column; gap: 10px;">
                        <div class="ui segment" :key="index" v-for="(button, index) in buttons"
                             style="padding: 10px; margin: 0;">
                            <div class="two fields">
                                <div class="field">
                                    <label>Type</label>
                                    <select v-model="button.type" class="ui dropdown"
                                            :aria-label="'button ' + index + ' type'">
                                        <option v-for="t in buttonTypes" :value="t.value">{{ t.text }}</option>
                                    </select>
                                </div>
                                <div class="field">
                                    <label>Title (max 20 chars)</label>
                                    <input v-model="button.title" type="text" maxlength="20"
                                           placeholder="Button label"
                                           :aria-label="'button ' + index + ' title'">
                                </div>
                            </div>
                            <div class="field" v-if="button.type === 'reply'">
                                <label>ID (optional, returned when pressed)</label>
                                <input v-model="button.id" type="text" placeholder="Defaults to the title"
                                       :aria-label="'button ' + index + ' id'">
                            </div>
                            <div class="field" v-if="button.type === 'cta_url'">
                                <label>URL</label>
                                <input v-model="button.url" type="text" placeholder="https://example.com"
                                       :aria-label="'button ' + index + ' url'">
                            </div>
                            <div class="field" v-if="button.type === 'cta_call'">
                                <label>Phone Number</label>
                                <input v-model="button.phone_number" type="text"
                                       placeholder="5588999999999"
                                       :aria-label="'button ' + index + ' phone'">
                            </div>
                            <div class="field" v-if="button.type === 'copy'">
                                <label>Code to copy</label>
                                <input v-model="button.copy_code" type="text"
                                       placeholder="Code copied when pressed"
                                       :aria-label="'button ' + index + ' copy code'">
                            </div>
                            <button class="mini ui red button" @click="deleteButton(index)" type="button">
                                <i class="minus circle icon"></i> Remove button
                            </button>
                        </div>
                        <div class="field">
                            <button class="mini ui primary button" @click="addButton" type="button"
                                    :disabled="buttons.length >= 3">
                                <i class="plus icon"></i> Button
                            </button>
                        </div>
                    </div>
                </div>
            </form>
        </div>
        <div class="actions">
            <button class="ui approve positive right labeled icon button" :class="{'loading': this.loading, 'disabled': !isValidForm() || loading}"
                 @click.prevent="handleSubmit">
                Send
                <i class="send icon"></i>
            </button>
        </div>
    </div>
`
}
