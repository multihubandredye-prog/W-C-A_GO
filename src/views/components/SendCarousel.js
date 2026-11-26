import FormRecipient from './generic/FormRecipient.js';

export default {
    name: 'SendCarousel',
    components: { FormRecipient },
    data() {
        return { phone: '', type: window.TYPEUSER, body: '', image_paths: '', button_text: '', card_titles: '', files: [], selection_error: '', loading: false };
    },
    computed: {
        imageLimit() { return window.CAROUSEL_MAX_IMAGE_BYTES; },
        paths() { return this.image_paths.split(',').map(path => path.trim()); },
        titles() { return this.card_titles.split(/\r?\n/).map(title => title.trim()); },
        validationMessage() {
            if (!this.phone.trim()) return 'Informe o destinatário.';
            if (!this.files.length && (this.paths.length < 2 || this.paths.length > 10 || this.paths.some(path => !path)))
                return 'Informe de 2 a 10 caminhos, separados por vírgulas.';
            if (!this.files.length && this.paths.some(path => !path.startsWith('/storage/emulated/0/')))
                return 'As imagens devem estar em /storage/emulated/0/ no aparelho da API.';
            if (this.files.length === 1) return 'Selecione pelo menos 2 imagens.';
            if (this.selection_error) return this.selection_error;
            if (Array.from(this.body.trim()).length > 1024 || Array.from(this.button_text.trim()).length > 20)
                return 'O texto da mensagem ou do botão excede o limite.';
            if (this.card_titles.trim() && (this.titles.length !== (this.files.length || this.paths.length) || this.titles.some(title => !title || Array.from(title).length > 160)))
                return 'Informe um título por imagem, sem linhas vazias e com até 160 caracteres cada.';
            return '';
        }
    },
    methods: {
        openModal() {
            $('#modalSendCarousel').modal({ onApprove: function () { return false; } }).modal('show');
        },
        addFiles(event) {
            const chosen = Array.from(event.target.files || []);
            event.target.value = '';
            if (!chosen.length) return;
            if (this.files.length + chosen.length > 10) {
                this.selection_error = 'Máximo de 10 imagens. A última seleção não foi adicionada.';
                return;
            }
            if (chosen.some(file => !/\.(jpe?g|png)$/i.test(file.name) || file.size <= 0 || file.size > window.CAROUSEL_MAX_IMAGE_BYTES)) {
                this.selection_error = 'Selecione arquivos JPEG ou PNG dentro do limite de imagem da API.';
                return;
            }
            this.selection_error = '';
            this.files.push(...chosen);
            this.image_paths = '';
        },
        removeFile(index) { this.files.splice(index, 1); this.selection_error = ''; },
        moveFile(index, direction) {
            const target = index + direction;
            if (target < 0 || target >= this.files.length) return;
            const item = this.files.splice(index, 1)[0]; this.files.splice(target, 0, item);
        },
        async handleSubmit() {
            if (this.validationMessage || this.loading) return;
            this.loading = true;
            let payload;
            if (this.files.length) {
                payload = new FormData();
                payload.append('phone', this.phone.trim() + this.type);
                this.files.forEach(file => payload.append('images', file, file.name));
                if (this.body.trim()) payload.append('body', this.body.trim());
                if (this.button_text.trim()) payload.append('button_text', this.button_text.trim());
                if (this.card_titles.trim()) payload.append('card_titles', JSON.stringify(this.titles));
            } else {
                payload = { phone: this.phone.trim() + this.type, image_paths: this.paths.join(',') };
                if (this.body.trim()) payload.body = this.body.trim();
                if (this.button_text.trim()) payload.button_text = this.button_text.trim();
                if (this.card_titles.trim()) payload.card_titles = this.titles;
            }
            try {
                const response = await window.http.post('/send/carousel', payload);
                window.showSuccessInfo(response.data.message);
                $('#modalSendCarousel').modal('hide');
                this.phone = ''; this.body = ''; this.image_paths = ''; this.button_text = ''; this.card_titles = ''; this.files = []; this.selection_error = '';
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
                    <label>Imagens (2 a 10)</label>
                    <input id="carousel-image-files" type="file" accept="image/jpeg,image/png" multiple style="display:none" @change="addFiles">
                    <label for="carousel-image-files" class="ui positive button" :class="{disabled: files.length >= 10}">
                        <i class="upload icon"></i> Adicionar imagens
                    </label>
                    <div style="margin: 10px 0">{{ files.length }} de 10 imagens selecionadas <span v-if="files.length === 10">— limite atingido</span></div>
                    <div v-if="selection_error" class="ui warning message">{{ selection_error }}</div>
                    <div v-for="(file, index) in files" :key="index" style="display:flex; align-items:center; gap:6px; margin:5px 0;">
                        <span style="flex:1; overflow-wrap:anywhere">{{ index + 1 }}. {{ file.name }}</span>
                        <button type="button" class="ui mini button" :disabled="index === 0" @click="moveFile(index, -1)" :aria-label="'Mover ' + file.name + ' para cima'">↑</button>
                        <button type="button" class="ui mini button" :disabled="index === files.length - 1" @click="moveFile(index, 1)" :aria-label="'Mover ' + file.name + ' para baixo'">↓</button>
                        <button type="button" class="ui mini red button" @click="removeFile(index)" :aria-label="'Remover ' + file.name">×</button>
                    </div>
                    <small>JPEG ou PNG, até {{ (imageLimit / 1000000).toLocaleString() }} MB e 25 megapixels por imagem. A ordem da lista define a ordem dos cartões. O total dos arquivos deve respeitar o limite geral da requisição.</small>
                </div>
                <div v-if="!files.length" class="field">
                    <label>Ou informe caminhos locais (2 a 10, separados por vírgulas)</label>
                    <textarea v-model="image_paths" rows="3" placeholder="/storage/emulated/0/imagens/foto1.jpg, /storage/emulated/0/imagens/foto2.png"></textarea>
                    <small>Use caminhos locais no aparelho onde a API está rodando, na ordem desejada. Não são aceitos links nem nomes com vírgula. JPEG/PNG: até {{ (imageLimit / 1000000).toLocaleString() }} MB e 25 megapixels por imagem.</small>
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
