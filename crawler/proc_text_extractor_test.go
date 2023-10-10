package crawler

import (
	"context"
	"fmt"
	"testing"
	"unicode/utf8"

	"github.com/odit-bit/invoker/internal/pipeline"
)

func Test_textExtractor(t *testing.T) {
	content := `<html>
	<head>
	<title>Test title</title>
	</head>
	<body>
	<div>Some<span> content</span></div>
	</body>
	</html>
	`

	p := new(payload)
	p.RawContent.WriteString(content)

	te := newTextExtractor()
	res, err := te.Process(context.TODO(), p)
	if err != nil {
		t.Error(err)
	}

	v := res.(*payload)
	if v.Title != "Test title" && v.TextContent != "Some content" {
		t.Errorf("title and content should same\ntitle:%v\t textContent:%v\n ", v.Title, v.TextContent)
	}
}

func Test_textExtractor2(t *testing.T) {
	content := "<html><head><title>Test title</title></head>	<body><div>Some<span> content</span></div><div>Some<span> Bj�rn H�hrmann � </span></div></body></html>"

	p := new(payload)
	p.RawContent.WriteString(content)

	te := newTextExtractor()
	res, err := te.Process(context.TODO(), p)
	if err != nil {
		t.Error(err)
	}

	v := res.(*payload)
	if !utf8.Valid([]byte(v.TextContent)) {
		t.Fatalf("not utf8: %v", v.TextContent)
	}

	if v.Title != "Test title" {
		t.Fatalf("got:%v\n", v.Title)
	}

	if v.TextContent != "Some contentSome Bjrn Hhrmann" {
		t.Fatalf("got:%v \n", v.TextContent)
	}
}

var rp pipeline.Payload

func Benchmark_text_extractor(b *testing.B) {
	html := []byte(fmt.Sprintf("<html><head><title>Test title</title></head>	<body><div>Some<span> content</span></div><div>Some<span>%v</span></div></body></html>", nonUTF8))

	p := new(payload)
	p.RawContent.Write(html)
	te := newTextExtractor()
	for i := 0; i < b.N; i++ {

		resultPayload, err := te.Process(context.TODO(), p)
		if err != nil {
			b.Fatal(err)
		}
		rp = resultPayload

	}
}

var nonUTF8 = `Nft, dai quadri al vino, fino ai campioni dello sport: così gli investimenti diventano «popolari»- Corriere.it desktop includes2013/SSI/notification/global.json /includes2013/SSI/utility/ajax_ssi_loader.shtml Home Opinioni Le Firme Lettere al direttore Il caffè di Gramellini Lo dico al Corriere Italians di Beppe Severgnini Il Twitter del direttore Padiglione Italia Letti da rifare di D'Avenia Facce nuove Datablog English / Chinese Editorials (English version) Chinese (Chinese Version) Video news Inchieste DocuWeb Contenuti Premium Sfoglia il giornale Extra Per Voi Rassegna Stampa Food Issue Firme di Corriere Cosa sto leggendo Politica Elezioni 2019 Elezioni Comunali 2018 Risultati Elezioni 2018 Risultati Elezione presidente Camera Risultati Elezione presidente Senato Elezioni Politiche 2018 Risultati Elezioni 2018 Risultati Elezione presidente Camera Risultati Elezione presidente Senato Lazio Elezioni Regionali 2018 Friuli Venezia Giulia Molise Lombardia Lazio Elezioni 2017 Elezioni Regionali Sicilia 2017 Referendum per l'autonomia Elezioni Comunali 2017 Primarie PD 2017 Speciali e Elezioni 2016 Referendum Costituzionale La Crisi di Governo Comunali 2016 Cronache Il crollo del ponte Morandi a Genova Royal Wedding Vajont, 50 anni dopo La strage del Mediterraneo Esteri Elezioni Midterm USA 2018 Risultati Le 25 sfide Elezioni Germania 2017 Elezioni Regno Unito 2017 Elezioni Presidenziali Francia 2017 Elezioni Presidenziali USA 2016 Economia Economia Finanza Borse e fondi Spread Principali indici Risparmio Guide Tasse Le vostre domande Calcolatori Consumi Casa Mutui Affitti Lavoro La nuovola del lavoro Guide Innovazione FinTech Intelligenza artificiale Start Up Agritech e Agrifood Pensioni Le vostre domande Calcolatori Guide Imprese Family Business Le storie Apri la tua impresa L'Economia del futuro L'Italia genera futuro Innovazione L'Italia che investe Opinioni Eventi Calcolatori Professionisti Ingegneri Avvocati Consulenti del Lavoro Commercialisti Partite Iva Moda Euractiv Nautica Sport Le dirette Serie A Calendario e risultati Le dirette Classifiche Marcatori Albo d'oro Videonews Serie B Calendario e risultati Le dirette Play-Off Play-Out Classifiche Marcatori Videonews Coppe Champions League Europa League Coppa Italia Nations League Calendario e risultati Le dirette Classifiche Formula 1 Le Dirette Calendario e Risultati Classifica Motomondiale Le Dirette Calendario e Risultati Classifica Barcolana50 Rugby Vela Corri, nuota, pedala Video sport Man of the match Formula 1 Pianeta 2030 Cultura 100 anni da Caporetto La Lettura Dataroom LiberiTutti Futura Buone Notizie 7 - Settimanale Il Lunghissimo Lungomare Scuola Elementari Medie Superiori Università Blog Speciale maturità Maturità 2017 La parola della settimana Spettacoli Prima della Scala 2018 Festival di Venezia Cannes 2018 OSCAR 2018 Sanremo 2019 Stasera in TV Film al Cinema YourVoice Le videorubriche TeleVisioni di A.Grasso Il Film di P.Mereghetti Mi Piace di M.L. Agnese Per niente Candida Tv Usa di S.Carini Festival di Sanremo Edizione 2018 Edizione 2017 Edizione 2016 Marilyn, il racconto del cinema Salute Cardiologia Diabete La dolce vita Dermatologia Speciale Psoriasi Disabilità Blog InVisibili Malattie Infettive Speciale influenza Malattie Rare Neuroscienze Speciale Sonno Nutrizione Ricette della salute Alimentazione sana Come conservare i cibi Ossa e Muscoli Pediatria I video del Pronto soccorso Blog Dubbi di mamma e papà Cannabis Reumatologia Che cosa sono i reumatismi I centri di cura in Italia Sportello cancro L'esperto risponde Esami del sangue I video di salute Webapp TUMORI CUORE Scienze Ambiente Settegreen Awards 2016 Forum Terra, Fuoco, Aria, Acqua Animali e dintorni Fotovoltaico ed eolico Animali Cani, gatti & co. Amici da salvare Il veterinario risponde Innovazione Login: Le domande che facciamo a Google Video tecnologia App & Soft Videogiochi Guide digitali Serie tv Guide I Blog Vita digitale Mal di tech 6 gradi eStory Hei Book Visioni Future Sentimeter L'Ora Legale Eliza Silicon Valley Piazza Digitale Archivio I Blog Vita digitale Mal di tech 6 gradi eStory Hei Book Visioni Future Sentimeter L'Ora Legale Eliza Silicon Valley Piazza Digitale Motori Video motori Speciale Ibrido Salone di Francoforte Salone di Ginevra Salone di Parigi 2018 Speciale Guido con un cane Prove Moto Tecnologia Auto d'epoca Anteprima Lifestyle Attualità Blog Heavy Rider Speciale Le vie dello sport Viaggi IN VIAGGIO CON CORRIERE Casa Eventi Salone del Mobile OROLOGI - Lo Speciale IL BELLO DELL'ITALIA COOK NEWS RICETTE WINE & COCKTAILS EVENTI LOCALI CIBO A REGOLA D'ARTE IoDonna Style 27Ora Il Tempo delle Donne 2018 La FELICITÀ. Adesso UOMINI - I segni del cambiamento Moda Speciale Natale News Beauty Sportswear Moda & Business Sfilate donna Sfilate uomo SOLFERINO LIBRI RISPARMI, MERCATI, IMPRESE ABBONATI Abbonati a 1€ AL MESE ABBONATI ORA Login Profilo Newsletter Abbonamento Logout L'ECONOMIA CORRIERE DELLA SERA X FINANZA BORSA E FONDI RISPARMIO TASSE CONSUMI CASA LAVORO INNOVAZIONE PENSIONI GUIDE IMPRESE MODA OPINIONI EVENTI PROFESSIONISTI EURACTIV Nautica Ecobonus Israele contro Hamas, le ultime notizie | le potenzialit� Nft, dai quadri al vino, fino ai campioni dello sport: cos� gli investimenti diventano �popolari� di Pieremilio Gadda Acquistare e vendere un solo metro quadro di un immobile di pregio, insieme ai diritti che ne derivano, come il reddito generato dal canone di locazione, pro quota. Custodire un paniere digitale di sneaker da collezione indossate dai pi� grandi atleti della storia recente. O ancora possedere una frazione di un’importante opera d’arte, altrimenti inaccessibile. Ecco un assaggio delle opportunit� che oggi si possono cogliere nel mondo degli investimenti alternativi, grazie all’innovazione tecnologica. Tutte si basano su applicazioni della blockchain: un registro contabile — distribuito su pi� nodi (computer), connessi tra loro — in cui vengono memorizzate le transazioni digitali in modo sicuro e immodificabile. �Ci� che espande le opportunit� della blockchain sono i nuovi smart contract, contratti digitali che possono codificare una serie di istruzioni, automatizzando i processi�, spiega Sandy Kaul, head of digital assets e Industry Advisory Services di Franklin Templeton. Ad esempio negli Stati Uniti il noto chef Tom Collichio ha lanciato una serie di pizze in nft (non fungible token, gettoni digitali), che includono speciali diritti: �chi possiede uno di questi nft entra a far parte di una community di persone che hanno accesso a eventi esclusivi, corsi, accessori per la cucina e cos� via. � il contratto digitale contenuto nell’nft a certificare che tu sia membro della community�. I numeri della finanza decentralizzataLa stessa tecnologia pu� rendere pi� liquido anche il mercato immobiliare: attraverso la tokenizzazione di un immobile, un grande edificio pu� essere suddiviso in frazioni digitali da scambiare su piattaforme dedicate. Ogni asset token certifica il possesso reale di una parte dell’immobile e dei diritti collegati, �che possono essere venduti a un altro investitore, rendendo molto pi� liquido un mercato di per s� illiquido�, osserva Kaul. Il quadro normativo � pronto ad accogliere queste novit�? �Oggi le opportunit� offerte dall’innovazione finanziaria sono pi� avanti rispetto ai regolatori — chiosa Kaul — anche se la cornice legale di riferimento non � ancora del tutto chiara. Ma il suo completamento avverr� molto velocemente, specialmente per i nuovi tipi di asset: basti pensare che in alcuni Paesi, come Cipro e la Svezia, la blockchain � gi� utilizzata per tenere il registro delle propriet� fondiarie�. Intanto, secondo un report di Grand View Research, il mondo della finanza decentralizzata ha gi� raggiunto un valore di 13,5 miliardi di dollari, nel 2022 e si prevede che possa espandersi a un tasso di crescita annuo composto del 46% tra oggi e il 2030. E non mancano i punti di convergenza con la finanza tradizionale. Alcuni progetti pilota, ad esempio, utilizzano la blockchain per rendere pi� efficiente l’industria dei fondi comuni d’investimento. �Oggi per le attivit� amministrative collegate ai fondi utilizziamo molti registri, uno per le sottoscrizioni, uno per i rimborsi, uno per i pagamenti e cos� via, la cui gestione presuppone elevati costi operativi. Grazie alla blockchain riusciamo a ridurli, rendendo pi� efficienti i processi. Abbiamo gi� avviato un progetto pilota�, dice la manager. Per approfondire: Investire per la laurea dei figli battendo l’inflazione: tre soluzioni a 5, 10 e 15 anni Simontacchi: �Innovazione e ricerca. Attrarre pi� investimenti per spingere la crescita� Piazza Affari, quando parte il cambio di stagione? I titoli da riscoprire per l’autunno Noi, robot. Nasce prima l’intuizione o la tecnologia che la rende possibile? Borse e Ipo, in settembre tornano le matricole (in prima fila la Birkenstock di Barbie) Intelligenza artificiale, 8,4 milioni di lavoratori a rischio: l’allarme di Confartigianato Wall Street, Cina, Europa: come investire nel secondo semestre 2023, i portafogli L’iniziativa made in UsaNel frattempo, negli Usa, Franklin Templeton ha lanciato un fondo di venture capital basato su blockchain. �Chi investe nei mercati privati riceve delle azioni. L’investimento viene monetizzato solo, ad esempio quando l’azienda viene quotata in Borsa. Nel nostro modello, invece, l’investitore non riceve equity, ma token, associati alle aziende, cedibili dopo 18 mesi. Se vuoi realizzare a pieno il valore dell’investimento, � necessario mantenere e gestire il paniere di token per molti anni e ad occuparsene � il gestore del fondo. Ma c’� la possibilit� di liquidare prima delle posizioni� Si tratta di una delle aree pi� interessanti su cui sta lavorando Franklin Templeton. �L’altra � lo sviluppo di un protocollo per la creazione di fondi specializzati sui cultural asset: beni o attivit� che possono suscitare emozioni nell’investitore: pensiamo a token per gli appassionati di Formula uno, che permettano di incontrare personalmente i piloti professionisti. O ancora a dei gettoni che consentano di avere accesso a bottiglie di vino pregiate, prodotte dalla propria cantina preferita: avremo sempre pi� prodotti strutturati per consegnare un ritorno emotivo, oltre che finanziario: metteranno i piccoli investitori nelle condizioni di accedere ad asset alternativi oggi riservati a operatori istituzionali. Oggi possiamo costruire portafogli con oltre 20 tipologie di token, non solo criptovalute�, conclude Kaul. Iscriviti alle newsletter di L'Economia Whatever it Takes di Federico Fubini Le sfide per l’economia e i mercati in un mondo instabile Europe Matters di Francesca Basso e Viviana Mazza L’Europa, gli Stati Uniti e l’Italia che contano, con le innovazioni e le decisioni importanti, ma anche le piccole storie di rilievo One More Thing di Massimo Sideri Dal mondo della scienza e dell’innovazione tecnologica le notizie che ci cambiano la vita (più di quanto crediamo) E non dimenticare le newsletter L'Economia Opinioni e L'Economia Ore 18 13 set 2023 © RIPRODUZIONE RISERVATA   Leggi i contributi SCRIVI ULTIME NOTIZIE DA L’ECONOMIA > il patto Trimestre anti-inflazione, da Ferrero a Mondelez: ecco i marchi che hanno aderito di Valentina Iorio consumi La Francia contro Amazon: tassa di 3 euro sulle spedizioni per difendere i librai di Redazione Economia Immobiliare I mutui delle case costano di più al sud, dove i tassi sono superiori al 6%: i dati per regione di Diana Cavalcoli legge di bilancio Manovra ristretta da 22 miliardi: caccia ai fondi per la sanità, le ultime ipotesi di Valentina Iorio Caro-vita Milano non è una città per giovani: due su tre spendono più di quanto guadagnano di Redazione Economia Conto corrente, come scegliere quello pi� adatto a noi? La nostra guida Le guide Le guide per approfondire i temi più discussi Leggi una guida Contratto d'affittoIl TirocinioPensioniMutuoModello 730 Vedi tutte le guide Product manager, in Italia il 50% delle aziende ne assumer� uno entro 3 anni di Massimiliano Jattoni Dall’As�n Monete da collezione, ora tocca a Italo Calvino per i 100 anni dalla nascita di Redazione Economia Berloni, Arredissima compra lo storico marchio di cucine: operazione da 2 milioni di Redazione Economia Spesa alimentare, costa di pi� e si compra di meno: lo certifica anche l’Istat di Michelangelo Borrillo Guardrail, a rischio 600 mila chilometri di strade italiane: l’allarme degli esperti di Redazione Economia Otk Kart con Vega: nasce il colosso mondiale (italiano) dei Kart da gara di Emily Capozucca Lago di Garda, navigazione in crescita: per la stagione 2023 +7,5% di passeggeri di Redazione Economia Bindi, cessione da 1 miliardo: il fondo Bc Partners mette in vendita il gruppo dolciario di Redazione Economia Fratelli Polli Spa: presentato il primo piano di sostenibilit� del gruppo di Valeriano Musiu Intesa Sanpaolo apre a Napoli un centro di innovazione con Talent Garden di Andrea Rinaldi Barcolana 55, pi� di 1.500 iscrizioni per la regata velica pi� grande del mondo di Antonio Macaluso Export, qual � l’andamento delle imprese italiane? A luglio calo del 7,7% di Redazione Economia Da Bitcoin a Dogecoin, la crypto-ricchezza degli italiani cala a 1,23 miliardi di Gabriele Petrucciani Airbnb, non solo affitti brevi, ora punta sul lungo periodo: come cambier� l’offerta di Alessia Conzonato Quanto costano cani e gatti? Il caro vita per gli animali pesa pi� dell’inflazione di Diana Cavalcoli Chi siamo | The Trust Project Abbonati a Corriere della Sera | Gazzetta | El Mundo | Marca | RCS Mediagroup | Fondazione Corriere | Fondazione Cutuli | Quimamme | OFFERTE CORRIERE STORE | Copyright 2023 © RCS Mediagroup S.p.a. Tutti i diritti sono riservati | Per la pubblicità: CAIRORCS MEDIA SpA - Direzione Pubblicit- Direzione Pubblicità RCS MediaGroup S.p.A. - Divisione Quotidiani Sede legale: via Angelo Rizzoli, 8 - 20132 Milano | Capitale sociale: Euro 270.000.000,00 Codice Fiscale, Partita I.V.A. e Iscrizione al Registro delle Imprese di Milano n.12086540155 | R.E.A. di Milano: 1524326 | ISSN 2499-0485 Servizi | Scrivi | Cookie policy e privacy | Preferenze sui Cookie`
