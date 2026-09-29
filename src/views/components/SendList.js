// export Vue Component
import FormRecipient from "./generic/FormRecipient.js";

export default {
    name: 'SendList',
    components: {
        FormRecipient
    },
    data() {
        return {
            phone: '',
            type: window.TYPEUSER,
            loading: false,
            description: '',
            title: '',
            button_text: '',
            footer: '',
            sections: [this.newSection()],
        }
    },
    computed: {
        phone_id() {
            return this.phone + this.type;
        }
    },
    methods: {
        newSection() {
            return { title: '', rows: [this.newRow()] };
        },
        newRow() {
            return { row_id: '', title: '', description: '' };
        },
        openModal() {
            $('#modalSendList').modal({
                onApprove: function () {
                    return false;
                }
            }).modal('show');
        },
        isValidForm() {
            if (this.type !== window.TYPESTATUS && !this.phone.trim()) {
                return false;
            }
            if (!this.description.trim()) {
                return false;
            }
            if (this.sections.length < 1) {
                return false;
            }
            return this.sections.every(section =>
                section.rows.length > 0 &&
                section.rows.every(row => row.title.trim() !== '')
            );
        },
        async handleSubmit() {
            if (!this.isValidForm() || this.loading) {
                return;
            }
            try {
                let response = await this.submitApi()
                window.showSuccessInfo(response)
                $('#modalSendList').modal('hide');
            } catch (err) {
                window.showErrorInfo(err)
            }
        },
        async submitApi() {
            this.loading = true;
            try {
                const payload = {
                    phone: this.phone_id,
                    description: this.description,
                    sections: this.sections.map(section => {
                        const item = {
                            rows: section.rows.map(row => {
                                const rowItem = { title: row.title };
                                if (row.row_id.trim()) rowItem.row_id = row.row_id;
                                if (row.description.trim()) rowItem.description = row.description;
                                return rowItem;
                            })
                        }
                        if (section.title.trim()) item.title = section.title;
                        return item;
                    })
                }
                if (this.title.trim()) payload.title = this.title;
                if (this.button_text.trim()) payload.button_text = this.button_text;
                if (this.footer.trim()) payload.footer = this.footer;

                const response = await window.http.post(`/send/list`, payload)
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
            this.description = '';
            this.title = '';
            this.button_text = '';
            this.footer = '';
            this.sections = [this.newSection()];
        },
        addSection() {
            this.sections.push(this.newSection())
        },
        deleteSection(index) {
            this.sections.splice(index, 1)
        },
        addRow(section) {
            section.rows.push(this.newRow())
        },
        deleteRow(section, index) {
            section.rows.splice(index, 1)
        }
    },
    template: `
    <div class="olive card" @click="openModal()" style="cursor: pointer">
        <div class="content">
            <a class="ui olive right ribbon label">Send</a>
            <div class="header">Send List</div>
            <div class="description">
                Send a selection list with sections and rows
            </div>
        </div>
    </div>

    <!--  Modal SendList  -->
    <div class="ui small modal" id="modalSendList">
        <i class="close icon"></i>
        <div class="header">
            Send List
        </div>
        <div class="content">
            <form class="ui form">
                <FormRecipient v-model:type="type" v-model:phone="phone"/>

                <div class="field">
                    <label>Description</label>
                    <textarea v-model="description" rows="2" placeholder="Main message text"
                              aria-label="list description"></textarea>
                </div>
                <div class="two fields">
                    <div class="field">
                        <label>Title (optional)</label>
                        <input v-model="title" type="text" placeholder="Header above the body"
                               aria-label="list title">
                    </div>
                    <div class="field">
                        <label>Button Text (optional)</label>
                        <input v-model="button_text" type="text" placeholder='Label of the button that opens the list'
                               aria-label="list button text">
                    </div>
                </div>
                <div class="field">
                    <label>Footer (optional)</label>
                    <input v-model="footer" type="text" placeholder="Small text at the bottom"
                           aria-label="list footer">
                </div>

                <div class="field">
                    <label>Sections</label>
                    <div style="display: flex; flex-direction: column; gap: 10px;">
                        <div class="ui segment" :key="sIndex" v-for="(section, sIndex) in sections"
                             style="padding: 10px; margin: 0;">
                            <div class="field">
                                <label>Section Title (optional)</label>
                                <input v-model="section.title" type="text" placeholder="Section heading"
                                       :aria-label="'section ' + sIndex + ' title'">
                            </div>
                            <div style="display: flex; flex-direction: column; gap: 5px;">
                                <div class="ui segment" :key="rIndex"
                                     v-for="(row, rIndex) in section.rows"
                                     style="padding: 10px; margin: 0; background: rgba(0,0,0,0.15);">
                                    <div class="two fields">
                                        <div class="field">
                                            <label>Row Title</label>
                                            <input v-model="row.title" type="text" maxlength="24"
                                                   placeholder="Item label"
                                                   :aria-label="'section ' + sIndex + ' row ' + rIndex + ' title'">
                                        </div>
                                        <div class="field">
                                            <label>Row ID (optional)</label>
                                            <input v-model="row.row_id" type="text" placeholder="Defaults to the title"
                                                   :aria-label="'section ' + sIndex + ' row ' + rIndex + ' id'">
                                        </div>
                                    </div>
                                    <div class="field">
                                        <label>Description (optional)</label>
                                        <input v-model="row.description" type="text" maxlength="72"
                                               placeholder="Secondary text under the title"
                                               :aria-label="'section ' + sIndex + ' row ' + rIndex + ' description'">
                                    </div>
                                    <button class="mini ui red button" @click="deleteRow(section, rIndex)"
                                            type="button">
                                        <i class="minus circle icon"></i> Remove row
                                    </button>
                                </div>
                            </div>
                            <div class="field" style="margin-top: 8px;">
                                <button class="mini ui primary button" @click="addRow(section)" type="button">
                                    <i class="plus icon"></i> Row
                                </button>
                                <button class="mini ui red button" @click="deleteSection(sIndex)" type="button"
                                        :disabled="sections.length <= 1">
                                    <i class="minus circle icon"></i> Remove section
                                </button>
                            </div>
                        </div>
                        <div class="field">
                            <button class="mini ui primary button" @click="addSection" type="button">
                                <i class="plus icon"></i> Section
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
