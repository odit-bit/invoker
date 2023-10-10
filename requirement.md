* link submission
    As an end user,
    I need to be able to submit new links to invoker,
    so as to update the link graph and make their contents searchable

    acceptance criteria:
        - Api endpoint will provided for submission
        - Submit link has criteria:
            1. must added to the graph
            2. must crawled by system and add to their index
        
        - already submitted link should receive by backend but not inserted twice.

* Search
    As an end user,
    I need to be able to submit FULL-TEXT SEARCH QUERY,
    so as to retreive a list of relevant matching result

    acceptance criteria:
        - Api endpoint will provided for full-text query
        - If the query matches multiple items, they are returned as a list that the end user can paginate through.
        - Each entry in the result list must contain the following items: 
            1. title or link description, the link to the content, and 
            2. a timestamp indicating when the link was last crawled. 

        - If feasible, the link may also contain a relevance score expressed as a percentage.
        - When the query does not match any item, an appropriate response should be returned to the end user.


* crawl link graph
    As the crawler backend system,
    I need to be able to obtain a list of sanitized links from the link graph,
    so as to fetch and index their contents while at the same time expanding
    the link graph with newly discovered links.

    acceptance criteria for this user story are as follows:
        - The crawler can query the link graph and receive a list of stale links that need to
        be crawled.
        - Links received by the crawler are retrieved from the remote hosts unless the remote server provides an ETag 
        or Last Modified header that the crawler has already seen before.
        - Retrieved content is scanned for links and the link graph gets updated.
        - Retrieved content is indexed and added to the search corpus.

* calculate PageRank scores
    As the PageRank calculator backend system,
    I need to be able to access the link graph,
    so as to calculate and persist the PageRank score for each link.

    The acceptance criteria for this user story are as follows:
        - The PageRank calculator can obtain an immutable snapshot of the entire link graph.
        - A PageRank score is assigned to every link in the graph.
        - The search corpus entries (body text) are annotated with the updated PageRank scores.


-.crawl service that can crawl the link to find new following link
    component that needed:
        -.link-store that can retrieve a portion of links and store it  back
        -.pipeline that can process link from src following existed link in page , extract the content


-.pageRank service that can ranking the link for index
    component that needed:
        -.link-store that can retrieve a portion of links and store it  back
        -.bspgraph ??

-.user-interface service that can store and retrieve the link from user


structure directory

Components:

    Role: 
        Components are individual building blocks or modules of a software system. 
        They represent the smaller, functional units of a system that perform specific tasks or functions.
    Responsibilities: 
        Each component has a well-defined set of responsibilities, and it encapsulates a particular piece of functionality. 
        Components can interact with each other to achieve more complex tasks.
    Example: 
        In a web application, components could include the user authentication module, database access layer, and user interface components like buttons and forms.


Domains:
    Role: 
        Domains, on the other hand, represent larger, cohesive areas of a software system. 
        They encapsulate related functionality and data that are grouped together due to their commonality.
    Responsibilities: 
        Domains are responsible for managing and organizing a specific set of related components. 
        They define the boundaries within which components operate and interact.
    Example: 
        In an e-commerce application, domains could include the product catalog domain, user management domain, and order processing domain. 
        Each domain would contain a set of components related to its specific functionality.