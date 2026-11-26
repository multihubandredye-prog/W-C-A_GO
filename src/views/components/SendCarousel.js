import FormRecipient from './generic/FormRecipient.js';

export default {
    name: 'SendCarousel',
    components: { FormRecipient },
    data() {
        return { phone: '', type: window.TYPEUSER, body: '', image_paths: '', button_text: '', card_titles: '', loading: false };
    },
    computed: {
        paths() { return this.image_paths.split(',').map(path => path.trim()); },
        titles() { return this.card_titles.split(/\r?\n/).map(title => title.trim()); },
        validationMessage() {
            if (!this.phone.trim()) return 'Informe o destinatário.';
            if (this.paths.length < 2 || this.paths.length > 10 || this.paths.some(path => !path))
                return 'Informe de 2 a 10 caminhos, separados por vírgulas.';
            if (this.paths.some(path => !path.startsWith('/storage/emulated/0/')))
                return 'As imagens devem estar em /storage/emulated/0/ no aparelho da API.';
            if (Array.from(this.body.trim()).length > 1024 || Array.from(this.button_text.trim()).length > 20)
                return 'O texto da mensagem ou do botão excede o limite.';
            if (this.card_titles.trim() && (this.titles.length !== this.paths.length || this.titles.some(title => !title || Array.from(title).length > 160)))
                return 'Informe um título por imagem, sem linhas vazias e com até 160 caracteres cada.';
            return '';
        }
    },
    methods: {
        openModal() {
            $('#modalSendCarousel').modal({ onApprove: function () { return false; } }).modal('show');
        },
        async handleSubmit() {
            if (this.validationMessage || this.loading) return;
            this.loading = true;
            const payload = { phone: this.phone.trim() + this.type, image_paths: this.paths.join(',') };
            if (this.body.trim()) payload.body = this.body.trim();
            if (this.button_text.trim()) payload.button_text = this.button_text.trim();
            if (this.card_titles.trim()) payload.card_titles = this.titles;
            try {
                const response = await window.http.post('/send/carousel', payload);
                window.showSuccessInfo(response.data.message);
                $('#modalSendCarousel').modal('hide');
                this.phone = ''; this.body = ''; this.image_paths = ''; this.button_text = ''; this.card_titles = '';
            } catch (error) {
                window.showErrorInfo(error.response?.data?.message || error.message || 'Erro ao enviar carrossel.');
            } finally {
                this.loading = false;
            }
        }
    },
    template: `
    <div class="purple card" @click="openModal()" style="cursor: pointer">
        <div class="content">
            <a class="ui purple right ribbon label">Send</a>
            <div class="header">Enviar carrossel</div>
            <div class="description">Envie de 2 a 10 imagens com botão de seleção.</div>
        </div>
    </div>
    <div class="ui small modal" id="modalSendCarousel">
        <i class="close icon"></i>
        <div class="header">Enviar carrossel de imagens</div>
        <div class="content" style="max-height: 70vh; overflow-y: auto;">
            <form class="ui form" @submit.prevent="handleSubmit">
                <FormRecipient v-model:type="type" v-model:phone="phone"/>
                <div class="field">
                    <label>Mensagem (opcional, até 1024 caracteres)</label>
                    <textarea v-model="body" rows="2" maxlength="1024" placeholder="Confira as imagens:"></textarea>
                </div>
                <div class="field">
                    <label>Caminhos das imagens (2 a 10, separados por vírgulas)</label>
                    <textarea v-model="image_paths" rows="3" placeholder="/storage/emulated/0/imagens/foto1.jpg, /storage/emulated/0/imagens/foto2.png"></textarea>
                    <small>Use caminhos locais no aparelho onde a API está rodando, na ordem desejada. Não são aceitos links nem nomes com vírgula. JPEG/PNG: até 5 MiB e 25 megapixels por imagem.</small>
                </div>
                <div class="field">
                    <label>Texto do botão (opcional, até 20 caracteres)</label>
                    <input v-model="button_text" maxlength="20" placeholder="Selecionar">
                </div>
                <div class="field">
                    <label>Títulos dos cartões (opcional, um por linha)</label>
                    <textarea v-model="card_titles" rows="3" placeholder="Primeiro cartão&#10;Segundo cartão"></textarea>
                    <small>Informe um título por imagem, na mesma ordem, ou deixe vazio para usar os nomes dos arquivos.</small>
                </div>
                <div v-if="validationMessage" class="ui small message">{{ validationMessage }}</div>
            </form>
        </div>
        <div class="actions">
            <button type="button" class="ui positive right labeled icon button" :class="{'loading': loading, 'disabled': validationMessage || loading}" :disabled="!!validationMessage || loading" @click="handleSubmit">
                Enviar <i class="send icon"></i>
            </button>
        </div>
    </div>`
};
