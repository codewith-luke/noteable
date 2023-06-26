import os
from dotenv import load_dotenv
from langchain import OpenAI
from langchain.chains.question_answering import load_qa_chain
from langchain.embeddings import OpenAIEmbeddings
from langchain.document_loaders import DirectoryLoader, UnstructuredMarkdownLoader
from langchain.text_splitter import RecursiveCharacterTextSplitter
from langchain.vectorstores import Chroma

load_dotenv()

api_key = os.getenv("OPENAI_API_KEY")
llm = OpenAI(openai_api_key=api_key, temperature=0, max_tokens=100)
embeddings = OpenAIEmbeddings(model="text-embedding-ada-002")


def insert_documents():
    loader = DirectoryLoader('./documents')
    loader_docs = loader.load()
    text_splitter = RecursiveCharacterTextSplitter(
        chunk_size=100,
        chunk_overlap=20,
        length_function=len,
    )

    for doc in loader_docs:
        docs = text_splitter.create_documents([doc.page_content])
        source = doc.metadata.get('source')

        for split_doc in docs:
            split_doc.metadata.update({'source': source})

        print("inserting", source)
        conn = Chroma.from_documents(
            docs,
            embeddings,
            collection_name="documents",
            persist_directory="./chroma_db",
        )

        conn.persist()


def index_document(path):
    file_name, file_extension = os.path.splitext(path)
    doc = None

    if file_extension == '.md':
        loader = UnstructuredMarkdownLoader(path)
        doc = loader.load()
        doc[0].metadata.update({'source': file_name})

    db = Chroma(collection_name="documents", persist_directory='chroma_db', embedding_function=embeddings)
    db.update_document()
    conn = Chroma.from_documents(
        doc,
        embeddings,
        collection_name="documents",
        persist_directory="./chroma_db",
    )

    conn.persist()


def query_documents(message):
    db = Chroma(collection_name="documents", persist_directory='chroma_db', embedding_function=embeddings)
    docs = db.similarity_search(message, k=1)
    chain = load_qa_chain(llm, chain_type="stuff")
    res = chain.run(input_documents=docs, question=message)
    print(res)


def find_document():
    db = Chroma(collection_name="documents", persist_directory='chroma_db', embedding_function=embeddings)
    docs = db.get(include=['metadatas'])
    sources = []

    for d in docs:
        source = d.metadata.get('source')
        if source not in sources:
            sources.append(source)

    print("found sources", sources)
    return sources


# query = "When is my dentist appointment? And on what day?"
# insert_documents()
# query_documents(query)
# index_document('./documents/test_document_3.md')
# doc = find_document(query)

